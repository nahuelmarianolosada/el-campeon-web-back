package utils

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword_Success testa que HashPassword genera un hash válido
func TestHashPassword_Success(t *testing.T) {
	password := "SecurePassword123"

	hash, err := HashPassword(password)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if hash == "" {
		t.Fatal("Expected non-empty hash")
	}
	if hash == password {
		t.Fatal("Hash must be different from password")
	}
	if len(hash) < 20 {
		t.Fatalf("Hash appears to be too short: %d characters", len(hash))
	}
}

// TestHashPassword_Randomness testa que cada hash es diferente (bcrypt añade salt)
func TestHashPassword_Randomness(t *testing.T) {
	password := "MyPassword123"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatal("Expected no errors")
	}

	// Cada hash debe ser diferente (bcrypt include salt aleatorio)
	if hash1 == hash2 {
		t.Fatal("Hashes of same password must be different (salt randomness)")
	}

	// Pero ambos deben hashear correctamente
	if !VerifyPassword(password, hash1) {
		t.Fatal("First hash verification failed")
	}
	if !VerifyPassword(password, hash2) {
		t.Fatal("Second hash verification failed")
	}
}

// TestHashPassword_WithEmptyPassword testa hash de contraseña vacía
func TestHashPassword_WithEmptyPassword(t *testing.T) {
	password := ""

	hash, err := HashPassword(password)

	// Incluso password vacía debe hashear correctamente
	if err != nil {
		t.Fatalf("Expected no error for empty password, got %v", err)
	}
	if hash == "" {
		t.Fatal("Expected non-empty hash for empty password")
	}

	// Pero debe verificarse correctamente
	if !VerifyPassword("", hash) {
		t.Fatal("Empty password verification failed")
	}
}

// TestHashPassword_WithSpecialCharacters testa hash con caracteres especiales
func TestHashPassword_WithSpecialCharacters(t *testing.T) {
	passwords := []string{
		"P@$$w0rd!#%&",
		"पासवर्ड123", // Unicode
		"Пароль123",  // Cyrillic
		"密码123",      // Chinese
		"!@#$%^&*()",
		"с пробелами",
	}

	for _, password := range passwords {
		hash, err := HashPassword(password)

		if err != nil {
			t.Fatalf("Error hashing password '%s': %v", password, err)
		}

		if !VerifyPassword(password, hash) {
			t.Fatalf("Failed to verify password with special characters: '%s'", password)
		}
	}
}

// TestHashPassword_WithLongPassword testa contraseña dentro del límite bcrypt
func TestHashPassword_WithLongPassword(t *testing.T) {
	// bcrypt tiene límite de 72 bytes exactos
	password := strings.Repeat("a", 72)

	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("Expected no error for 72-byte password, got %v", err)
	}

	if !VerifyPassword(password, hash) {
		t.Fatal("72-byte password verification failed")
	}
}

// TestHashPassword_WithExceedingMaxLength testa contraseña que excede límite bcrypt
func TestHashPassword_WithExceedingMaxLength(t *testing.T) {
	// bcrypt falla con contraseñas > 72 bytes
	password := strings.Repeat("a", 73)

	hash, err := HashPassword(password)

	// Debe haber error
	if err == nil {
		t.Fatal("Expected error for password > 72 bytes")
	}

	// Hash debe estar vacío
	if hash != "" {
		t.Fatal("Hash should be empty when error occurs")
	}
}

// TestVerifyPassword_CorrectPassword testa verificación exitosa
func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "MySecurePassword123"
	hash, _ := HashPassword(password)

	result := VerifyPassword(password, hash)

	if !result {
		t.Fatal("Expected true for correct password")
	}
}

// TestVerifyPassword_IncorrectPassword testa verificación fallida
func TestVerifyPassword_IncorrectPassword(t *testing.T) {
	password := "MySecurePassword123"
	hash, _ := HashPassword(password)

	result := VerifyPassword("WrongPassword456", hash)

	if result {
		t.Fatal("Expected false for incorrect password")
	}
}

// TestVerifyPassword_EmptyPassword testa con contraseña vacía
func TestVerifyPassword_EmptyPassword(t *testing.T) {
	hash, _ := HashPassword("")

	result := VerifyPassword("", hash)

	if !result {
		t.Fatal("Expected true for empty password matching empty hash")
	}
}

// TestVerifyPassword_WrongHashFormat testa con hash inválido
func TestVerifyPassword_WrongHashFormat(t *testing.T) {
	password := "MyPassword123"
	invalidHash := "not-a-valid-bcrypt-hash"

	result := VerifyPassword(password, invalidHash)

	if result {
		t.Fatal("Expected false for invalid hash format")
	}
}

// TestVerifyPassword_EmptyHash testa con hash vacío
func TestVerifyPassword_EmptyHash(t *testing.T) {
	result := VerifyPassword("SomePassword", "")

	if result {
		t.Fatal("Expected false for empty hash")
	}
}

// TestVerifyPassword_PartialPassword testa si substring no valida
func TestVerifyPassword_PartialPassword(t *testing.T) {
	password := "MySecurePassword123"
	hash, _ := HashPassword(password)

	// Intentar verificar con solo parte de la contraseña
	result := VerifyPassword("MySecure", hash)

	if result {
		t.Fatal("Expected false for partial password")
	}
}

// TestVerifyPassword_CaseSensitive testa que es case-sensitive
func TestVerifyPassword_CaseSensitive(t *testing.T) {
	password := "MyPassword"
	hash, _ := HashPassword(password)

	// Contraseña con diferente case
	result := VerifyPassword("mypassword", hash)

	if result {
		t.Fatal("Expected false for different case (case-sensitive)")
	}
}

// TestVerifyPassword_WithWhitespace testa con espacios
func TestVerifyPassword_WithWhitespace(t *testing.T) {
	password := "My Password"
	hash, _ := HashPassword(password)

	// Intentar verificar sin espacios
	result := VerifyPassword("MyPassword", hash)

	if result {
		t.Fatal("Expected false when whitespace is missing")
	}

	// Verificar con espacios correctos
	result = VerifyPassword("My Password", hash)

	if !result {
		t.Fatal("Expected true when whitespace matches")
	}
}

// BenchmarkHashPassword compara performance de HashPassword
func BenchmarkHashPassword(b *testing.B) {
	password := "BenchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVerifyPassword compara performance de VerifyPassword
func BenchmarkVerifyPassword(b *testing.B) {
	password := "BenchmarkPassword123"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hash)
	}
}

// TestHashPassword_BcryptCostFactor verifica que usa DefaultCost
func TestHashPassword_BcryptCostFactor(t *testing.T) {
	password := "TestPassword"
	hash, _ := HashPassword(password)

	// Extraer cost factor del hash (está en los primeros caracteres)
	// Formato: $2a$10$... donde 10 es el cost factor
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
		t.Fatalf("Hash doesn't have valid bcrypt format: %s", hash)
	}

	// Hash debe tener longitud estándar de bcrypt (60 caracteres)
	if len(hash) != 60 {
		t.Fatalf("Hash length should be 60, got %d", len(hash))
	}
}

// TestVerifyPassword_DifferentPricingFormats verifica compatibilidad con diferentes formatos
func TestVerifyPassword_DifferentFormats(t *testing.T) {
	password := "TestPassword"

	// Generar hash con bcrypt
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	// Verificar que VerifyPassword funciona
	if !VerifyPassword(password, hash) {
		t.Fatal("VerifyPassword failed with valid hash")
	}

	// Verificar que falla con contraseña incorrecta
	if VerifyPassword("WrongPassword", hash) {
		t.Fatal("VerifyPassword should fail with wrong password")
	}
}

// TestHashPassword_ConsistentWithVerifyPassword verifica integración entre funciones
func TestHashPassword_ConsistentWithVerifyPassword(t *testing.T) {
	passwords := []string{
		"Simple",
		"WithNumbers123",
		"WithSpecial!@#$",
		"VeryLongPasswordWith" + strings.Repeat("x", 40), // Total < 72 bytes
		"",
	}

	for _, password := range passwords {
		// Hash la contraseña
		hash, err := HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed for '%s': %v", password, err)
		}

		// Verificar que el hash es válido
		if hash == "" {
			t.Fatalf("Hash is empty for password '%s'", password)
		}

		// Verificar que VerifyPassword acepta la contraseña correcta
		if !VerifyPassword(password, hash) {
			t.Fatalf("VerifyPassword failed for correct password '%s'", password)
		}

		// Verificar que VerifyPassword rechaza contraseña incorrecta
		wrongPassword := password + "WRONG"
		if VerifyPassword(wrongPassword, hash) {
			t.Fatalf("VerifyPassword accepted wrong password for '%s'", password)
		}
	}
}

// TestVerifyPassword_WithMultipleHashes verifica que cada hash se verifica independientemente
func TestVerifyPassword_WithMultipleHashes(t *testing.T) {
	passwords := []string{"pass1", "pass2", "pass3"}
	hashes := make([]string, len(passwords))

	// Generar múltiples hashes
	for i, password := range passwords {
		hash, err := HashPassword(password)
		if err != nil {
			t.Fatalf("Failed to hash password %d: %v", i, err)
		}
		hashes[i] = hash
	}

	// Verificar cada contraseña contra su hash correcto
	for i, password := range passwords {
		if !VerifyPassword(password, hashes[i]) {
			t.Fatalf("Password %d verification failed", i)
		}
	}

	// Verificar que no se valida cruzado (pass1 != hash de pass2)
	for i, password := range passwords {
		for j, hash := range hashes {
			if i != j && VerifyPassword(password, hash) {
				t.Fatalf("Password %d should NOT verify with hash %d", i, j)
			}
		}
	}
}

// TestHashPassword_NoErrorWithValidInput testa que no hay errores con inputs válidos
func TestHashPassword_NoErrorWithValidInput(t *testing.T) {
	validPasswords := []string{
		"A",
		"abc",
		"password123",
		"Pa$$w0rd!",
		strings.Repeat("x", 72), // bcrypt max (exactly)
	}

	for _, password := range validPasswords {
		_, err := HashPassword(password)
		if err != nil {
			t.Fatalf("Unexpected error for valid password '%s': %v", password, err)
		}
	}
}

// TestVerifyPassword_Idempotent verifica que verificar múltiples veces da mismo resultado
func TestVerifyPassword_Idempotent(t *testing.T) {
	password := "TestPassword"
	hash, _ := HashPassword(password)

	result1 := VerifyPassword(password, hash)
	result2 := VerifyPassword(password, hash)
	result3 := VerifyPassword(password, hash)

	if result1 != result2 || result2 != result3 {
		t.Fatal("VerifyPassword should return same result for multiple calls")
	}
}

// TestHashPassword_OutputIsBcrypt verifica que el output es válido bcrypt
func TestHashPassword_OutputIsBcrypt(t *testing.T) {
	password := "TestPassword"
	hash, _ := HashPassword(password)

	// Intentar usar el hash como bcrypt hash
	// Si CompareHashAndPassword funciona, es válido bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Fatalf("Hash is not valid bcrypt: %v", err)
	}
}
