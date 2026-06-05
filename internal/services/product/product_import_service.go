package product

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ImportAction describes the resolved action for an import row.
type ImportAction string

const (
	ImportActionCreate ImportAction = "create"
	ImportActionUpdate ImportAction = "update"
	ImportActionError  ImportAction = "error"
)

// ImportRow is a single parsed and resolved entry from the file.
// Pointer fields are nil when the spreadsheet column was missing or empty,
// so we can distinguish "no change" from "set to zero/empty".
type ImportRow struct {
	LineNumber      int          `json:"line_number"`
	Action          ImportAction `json:"action"`
	SKU             string       `json:"sku"`
	Name            *string      `json:"name,omitempty"`
	Description     *string      `json:"description,omitempty"`
	Category        *string      `json:"category,omitempty"`
	PriceRetail     *float64     `json:"price_retail,omitempty"`
	PriceWholesale  *float64     `json:"price_wholesale,omitempty"`
	Stock           *int         `json:"stock,omitempty"`
	MinBulkQuantity *int         `json:"min_bulk_quantity,omitempty"`
	ImageURLs       *[]string    `json:"image_urls,omitempty"`
	IsActive        *bool        `json:"is_active,omitempty"`
	ExistingID      uint         `json:"existing_id,omitempty"`
	Errors          []string     `json:"errors,omitempty"`
}

// ImportResult is the response body for both preview and commit calls.
type ImportResult struct {
	DryRun   bool        `json:"dry_run"`
	Total    int         `json:"total"`
	ToCreate int         `json:"to_create"`
	ToUpdate int         `json:"to_update"`
	Errors   int         `json:"errors"`
	Created  int         `json:"created,omitempty"`
	Updated  int         `json:"updated,omitempty"`
	Skipped  int         `json:"skipped,omitempty"`
	Rows     []ImportRow `json:"rows"`
}

// ImportProducts parses an .xlsx or .csv file and either previews (dryRun=true)
// or applies (dryRun=false) the create/update operations, matching by SKU.
func (s *productService) ImportProducts(file io.Reader, fileName string, dryRun bool) (*ImportResult, error) {
	log.Printf("[productService.ImportProducts] INFO: Starting import - fileName=%s, dryRun=%v", fileName, dryRun)

	rawRows, err := parseImportFile(file, fileName)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{DryRun: dryRun, Rows: make([]ImportRow, 0, len(rawRows))}
	seenSKU := map[string]int{}

	for _, raw := range rawRows {
		row := raw

		if row.SKU == "" {
			row.Errors = append(row.Errors, "sku es obligatorio")
		} else {
			key := strings.ToLower(row.SKU)
			if prevLine, dup := seenSKU[key]; dup {
				row.Errors = append(row.Errors, fmt.Sprintf("SKU duplicado en el archivo (también en la fila %d)", prevLine))
			} else {
				seenSKU[key] = row.LineNumber
			}
		}

		if len(row.Errors) == 0 {
			existing, lookupErr := s.productRepo.FindBySKU(row.SKU)
			if lookupErr == nil && existing != nil {
				row.ExistingID = existing.ID
				row.Action = ImportActionUpdate
			} else if errors.Is(lookupErr, gorm.ErrRecordNotFound) || existing == nil {
				row.Action = ImportActionCreate
			} else {
				row.Errors = append(row.Errors, fmt.Sprintf("error consultando SKU: %v", lookupErr))
			}
		}

		validateImportRow(&row)

		if len(row.Errors) > 0 {
			row.Action = ImportActionError
			result.Errors++
		} else if row.Action == ImportActionCreate {
			result.ToCreate++
		} else if row.Action == ImportActionUpdate {
			result.ToUpdate++
		}
		result.Total++
		result.Rows = append(result.Rows, row)
	}

	if dryRun {
		log.Printf("[productService.ImportProducts] INFO: Preview ready - total=%d, toCreate=%d, toUpdate=%d, errors=%d", result.Total, result.ToCreate, result.ToUpdate, result.Errors)
		return result, nil
	}

	for i := range result.Rows {
		row := &result.Rows[i]
		switch row.Action {
		case ImportActionError:
			result.Skipped++
		case ImportActionCreate:
			req := buildCreateRequest(row)
			if _, err := s.CreateProduct(req); err != nil {
				row.Errors = append(row.Errors, err.Error())
				row.Action = ImportActionError
				result.Errors++
				result.Skipped++
				continue
			}
			result.Created++
		case ImportActionUpdate:
			req := buildUpdateRequest(row)
			if _, err := s.UpdateProduct(row.ExistingID, req); err != nil {
				row.Errors = append(row.Errors, err.Error())
				row.Action = ImportActionError
				result.Errors++
				result.Skipped++
				continue
			}
			result.Updated++
		}
	}
	log.Printf("[productService.ImportProducts] INFO: Import finished - created=%d, updated=%d, skipped=%d, errors=%d", result.Created, result.Updated, result.Skipped, result.Errors)
	return result, nil
}

func parseImportFile(file io.Reader, fileName string) ([]ImportRow, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".xlsx":
		return parseXLSX(file)
	case ".csv":
		return parseCSV(file)
	default:
		return nil, fmt.Errorf("formato no soportado: %q. Use .xlsx o .csv", ext)
	}
}

func parseCSV(file io.Reader) ([]ImportRow, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error leyendo CSV: %w", err)
	}
	return recordsToRows(records)
}

func parseXLSX(file io.Reader) ([]ImportRow, error) {
	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("error abriendo XLSX: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("el archivo no tiene hojas")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("error leyendo hoja %q: %w", sheets[0], err)
	}
	return recordsToRows(rows)
}

// headerAliases maps user-facing header names (Spanish/English) to canonical keys.
var headerAliases = map[string]string{
	"sku":                       "sku",
	"name":                      "name",
	"nombre":                    "name",
	"description":               "description",
	"descripcion":               "description",
	"descripción":               "description",
	"category":                  "category",
	"categoria":                 "category",
	"categoría":                 "category",
	"price_retail":              "price_retail",
	"precio_minorista":          "price_retail",
	"precio minorista":          "price_retail",
	"price_wholesale":           "price_wholesale",
	"precio_mayorista":          "price_wholesale",
	"precio mayorista":          "price_wholesale",
	"stock":                     "stock",
	"min_bulk_quantity":         "min_bulk_quantity",
	"cantidad_minima_mayorista": "min_bulk_quantity",
	"cantidad mínima mayorista": "min_bulk_quantity",
	"images":                    "images",
	"imagenes":                  "images",
	"imágenes":                  "images",
	"image_urls":                "images",
	"is_active":                 "is_active",
	"activo":                    "is_active",
}

func normalizeHeaders(headers []string) map[string]int {
	out := map[string]int{}
	for i, h := range headers {
		h = strings.TrimSpace(strings.ToLower(h))
		if h == "" {
			continue
		}
		if canonical, ok := headerAliases[h]; ok {
			out[canonical] = i
		}
	}
	return out
}

func recordsToRows(records [][]string) ([]ImportRow, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("el archivo está vacío")
	}
	headers := normalizeHeaders(records[0])
	if _, ok := headers["sku"]; !ok {
		return nil, fmt.Errorf("falta la columna obligatoria 'sku'")
	}

	rows := make([]ImportRow, 0, len(records)-1)
	for i, rec := range records[1:] {
		if isEmptyRow(rec) {
			continue
		}
		row := ImportRow{LineNumber: i + 2}

		raw := func(col string) (string, bool) {
			idx, ok := headers[col]
			if !ok || idx >= len(rec) {
				return "", false
			}
			v := strings.TrimSpace(rec[idx])
			return v, v != ""
		}

		if v, ok := raw("sku"); ok {
			row.SKU = v
		}
		if v, ok := raw("name"); ok {
			row.Name = strPtr(v)
		}
		if v, ok := raw("description"); ok {
			row.Description = strPtr(v)
		}
		if v, ok := raw("category"); ok {
			row.Category = strPtr(v)
		}
		if v, ok := raw("price_retail"); ok {
			f, err := parseDecimal(v)
			if err != nil {
				row.Errors = append(row.Errors, fmt.Sprintf("price_retail inválido: %q", v))
			} else {
				row.PriceRetail = &f
			}
		}
		if v, ok := raw("price_wholesale"); ok {
			f, err := parseDecimal(v)
			if err != nil {
				row.Errors = append(row.Errors, fmt.Sprintf("price_wholesale inválido: %q", v))
			} else {
				row.PriceWholesale = &f
			}
		}
		if v, ok := raw("stock"); ok {
			n, err := strconv.Atoi(v)
			if err != nil {
				row.Errors = append(row.Errors, fmt.Sprintf("stock inválido: %q", v))
			} else {
				row.Stock = &n
			}
		}
		if v, ok := raw("min_bulk_quantity"); ok {
			n, err := strconv.Atoi(v)
			if err != nil {
				row.Errors = append(row.Errors, fmt.Sprintf("min_bulk_quantity inválido: %q", v))
			} else {
				row.MinBulkQuantity = &n
			}
		}
		if v, ok := raw("images"); ok {
			parts := strings.Split(v, "|")
			urls := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					urls = append(urls, p)
				}
			}
			row.ImageURLs = &urls
		}
		if v, ok := raw("is_active"); ok {
			b := parseBool(v)
			row.IsActive = &b
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func isEmptyRow(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func parseDecimal(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	dotCount := strings.Count(s, ".")
	commaCount := strings.Count(s, ",")
	switch {
	case commaCount > 0 && dotCount == 0:
		// Spanish format: comma as decimal separator.
		s = strings.ReplaceAll(s, ",", ".")
	case commaCount > 0 && dotCount > 0:
		// Treat comma as thousands separator.
		s = strings.ReplaceAll(s, ",", "")
	}
	return strconv.ParseFloat(s, 64)
}

func parseBool(s string) bool {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "1", "true", "yes", "y", "si", "sí", "x", "verdadero":
		return true
	}
	return false
}

func validateImportRow(row *ImportRow) {
	if row.Action == ImportActionCreate {
		if row.Name == nil || strings.TrimSpace(*row.Name) == "" {
			row.Errors = append(row.Errors, "name es obligatorio para crear")
		}
		if row.Category == nil || strings.TrimSpace(*row.Category) == "" {
			row.Errors = append(row.Errors, "category es obligatorio para crear")
		}
		if row.PriceRetail == nil || *row.PriceRetail <= 0 {
			row.Errors = append(row.Errors, "price_retail debe ser mayor a 0 para crear")
		}
		if row.PriceWholesale == nil || *row.PriceWholesale <= 0 {
			row.Errors = append(row.Errors, "price_wholesale debe ser mayor a 0 para crear")
		}
		if row.Stock == nil {
			zero := 0
			row.Stock = &zero
		}
	}
	if row.Stock != nil && *row.Stock < 0 {
		row.Errors = append(row.Errors, "stock no puede ser negativo")
	}
	if row.PriceRetail != nil && *row.PriceRetail < 0 {
		row.Errors = append(row.Errors, "price_retail no puede ser negativo")
	}
	if row.PriceWholesale != nil && *row.PriceWholesale < 0 {
		row.Errors = append(row.Errors, "price_wholesale no puede ser negativo")
	}
}

func buildCreateRequest(row *ImportRow) *models.CreateProductRequest {
	req := &models.CreateProductRequest{SKU: row.SKU}
	if row.Name != nil {
		req.Name = *row.Name
	}
	if row.Description != nil {
		req.Description = *row.Description
	}
	if row.Category != nil {
		req.Category = *row.Category
	}
	if row.PriceRetail != nil {
		req.PriceRetail = *row.PriceRetail
	}
	if row.PriceWholesale != nil {
		req.PriceWholesale = *row.PriceWholesale
	}
	if row.Stock != nil {
		req.Stock = *row.Stock
	}
	if row.MinBulkQuantity != nil {
		req.MinBulkQuantity = *row.MinBulkQuantity
	}
	if row.ImageURLs != nil {
		images := make([]models.ProductImage, 0, len(*row.ImageURLs))
		for i, u := range *row.ImageURLs {
			images = append(images, models.ProductImage{ImageURL: u, DisplayOrder: i})
		}
		req.ImageURLs = images
	}
	return req
}

func buildUpdateRequest(row *ImportRow) *models.UpdateProductRequest {
	req := &models.UpdateProductRequest{
		Name:            row.Name,
		Description:     row.Description,
		Category:        row.Category,
		PriceRetail:     row.PriceRetail,
		PriceWholesale:  row.PriceWholesale,
		Stock:           row.Stock,
		MinBulkQuantity: row.MinBulkQuantity,
		IsActive:        row.IsActive,
	}
	if row.ImageURLs != nil {
		images := make([]models.ProductImage, 0, len(*row.ImageURLs))
		for i, u := range *row.ImageURLs {
			images = append(images, models.ProductImage{ImageURL: u, DisplayOrder: i})
		}
		req.ImageURLs = &images
	}
	return req
}

func strPtr(s string) *string { return &s }
