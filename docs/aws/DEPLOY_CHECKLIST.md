# 📋 Checklist Pre-Deploy - Sistema de Checkout Mejorado

## ✅ Backend Go - Verificaciones

### Código
- [ ] Compilación sin errores
  ```bash
  go build -o main ./cmd/main.go
  ```
- [ ] Tests pasando
  ```bash
  go test ./internal/services/payment/...
  go test ./internal/services/order/...
  ```
- [ ] Imports correctos (preferenceMp en lugar de orderMp)
- [ ] No hay código commented
- [ ] Logs configurados correctamente

### Archivos Modificados
- [ ] `internal/models/order.go` - delivery_method agregado
- [ ] `internal/models/payment.go` - payment_method enum actualizado
- [ ] `internal/services/payment/payment_service.go` - Usa Preference
- [ ] `internal/services/payment/mercadopago_client.go` - CreatePreference
- [ ] `internal/services/order/order_service.go` - Maneja delivery_method

### Archivos Nuevos
- [ ] `migrations/002_add_delivery_and_payment_methods.sql` - Existe

## ✅ Base de Datos - Verificaciones

### Migración
- [ ] Archivo SQL existe en `migrations/`
- [ ] Syntax correcto (probar con MySQL localmente)
- [ ] Backup de DB realizado antes de aplicar

### Aplicar Migración
```bash
# Verificar estructura actual
mysql -u usuario -p base_datos -e "DESCRIBE orders;"
mysql -u usuario -p base_datos -e "DESCRIBE payments;"

# Aplicar migración
mysql -u usuario -p base_datos < migrations/002_add_delivery_and_payment_methods.sql

# Verificar cambios
mysql -u usuario -p base_datos -e "DESCRIBE orders;" # debe mostrar delivery_method
mysql -u usuario -p base_datos -e "DESCRIBE payments;" # debe mostrar nuevos enums
```

### Validaciones Post-Migración
- [ ] Columna `delivery_method` existe en `orders`
- [ ] ENUM de `payment_method` actualizado en `payments`
- [ ] Índice en `delivery_method` creado
- [ ] No hay errores de integridad referencial

## ✅ API - Verificaciones

### Endpoints Nuevos/Modificados
- [ ] POST /api/orders - Acepta `delivery_method` (requerido)
- [ ] POST /api/payments - Acepta `payment_method` (requerido)
- [ ] GET /api/orders/:id - Retorna `delivery_method`
- [ ] GET /api/payments/order/:orderId - Funciona correctamente

### Validaciones
- [ ] Order: delivery_method == válido (soneof)
- [ ] Payment: payment_method == válido (oneof)
- [ ] Payment: amount == order.total (validación)
- [ ] Payment: orden no está cancelada

### Response Format
```bash
# Test delivery_method en order
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/orders/1 | jq .delivery_method

# Test payment_method en payment
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/payments/1 | jq .payment_method
```

## ✅ MercadoPago - Verificaciones

### Configuración
- [ ] Token de MercadoPago válido en config
- [ ] Variable de entorno configurada: `MERCADOPAGO_ACCESS_TOKEN`
- [ ] Token testeable (conexión a API es exitosa)

### Integración
- [ ] Usando `preference` en lugar de `order` SDK
- [ ] Preference creada correctamente
- [ ] Preference ID retornado en payment response
- [ ] URLs de preference generadas correctamente

### Testing
```bash
# Curl para crear payment con MP
curl -X POST http://localhost:8080/api/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": 1,
    "amount": 1210.00,
    "payment_method": "MP_CARD"
  }' | jq .mercadopago_preference_id

# Debe retornar un ID válido
```

## ✅ Frontend React - Verificaciones

### Archivos
- [ ] Checkout component copiado a `src/app/(authenticated)/checkout/page.tsx`
- [ ] API helpers copiados a `lib/api.ts`
- [ ] Imports correctos en componente

### Configuración
- [ ] `.env.local` tiene `NEXT_PUBLIC_API_URL`
- [ ] URL apunta al backend correcto
- [ ] Contextos `useCart()` y `useAuth()` funcionan

### Funcionalidad
- [ ] Selector de delivery_method funciona
- [ ] Selector de payment_method funciona
- [ ] Dirección se oculta/muestra según delivery_method
- [ ] Resumen muestra información correcta

### Testing Manual
- [ ] Seleccionar "Envío a domicilio"
- [ ] Llenar dirección válida
- [ ] Seleccionar método de pago
- [ ] Hacer clic en "Confirmar Pedido"
- [ ] ¿Redirige a MercadoPago o muestra código?

## ✅ Documentación - Verificaciones

- [ ] `CHECKOUT_INTEGRATION.md` - Existe y está completo
- [ ] `CHECKOUT_INTEGRATION_FRONTEND.md` - Existe y es claro
- [ ] `CHECKOUT_API_HELPERS.ts` - Tiene ejemplos de uso
- [ ] `API_EXAMPLES.sh` - Tiene ejemplos curl completos
- [ ] `IMPLEMENTATION_SUMMARY_CHECKOUT.md` - Explica cambios
- [ ] `QUICK_REFERENCE.md` - Resumen rápido disponible

## ✅ Seguridad - Verificaciones

### Validaciones
- [ ] Monto no puede ser modificado (validado en backend)
- [ ] Usuario no puede cambiar ID de orden
- [ ] Métodos de pago/entrega restrictivos (enums)
- [ ] JWT token requerido en endpoints protegidos

### Tests de Seguridad
```bash
# Test 1: Intentar con monto diferente
curl -X POST http://localhost:8080/api/payments \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"order_id": 1, "amount": 500.00, "payment_method": "MP_CARD"}'
# Debe fallar: "payment amount does not match"

# Test 2: Intentar con payment_method inválido
curl -X POST http://localhost:8080/api/payments \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"order_id": 1, "amount": 1210.00, "payment_method": "INVALID"}'
# Debe fallar: validación de enum

# Test 3: Intentar sin token
curl -X POST http://localhost:8080/api/payments \
  -d '{"order_id": 1, "amount": 1210.00, "payment_method": "MP_CARD"}'
# Debe fallar: 401 Unauthorized
```

## ✅ Performance - Verificaciones

- [ ] Índice en `delivery_method` creado
- [ ] Índice en `payment_method` creado
- [ ] No hay N+1 queries
- [ ] Queries optimizadas en repositories

## ✅ Logs y Debugging - Verificaciones

- [ ] Logs de payments creados/actualizados
- [ ] Logs de MercadoPago API calls
- [ ] Error messages descriptivos
- [ ] Debugging info disponible en dev

## ✅ Staging - Verificaciones

### Antes de Producción
1. [ ] Deploy a staging exitoso
2. [ ] DB staging migrada correctamente
3. [ ] Testing completo en staging
4. [ ] No hay errores en logs de staging
5. [ ] Performance aceptable

### Tests End-to-End en Staging
- [ ] Crear carrito con productos
- [ ] Ir a checkout
- [ ] Seleccionar entrega y pago
- [ ] Completar compra
- [ ] Verificar orden creada
- [ ] Verificar pago creado
- [ ] Verificar respuesta de API

## ✅ Producción - Verificaciones Pre-Deploy

### Final Checklist
1. [ ] Backup de DB de producción realizado
2. [ ] Rollback plan documentado
3. [ ] Team notificado de deploy
4. [ ] Feature flags configurados (si aplica)
5. [ ] Monitoring configurado

### Monitoreo Post-Deploy
1. [ ] Revisar logs cada 5 minutos durante 30 minutos
2. [ ] Verificar que nuevos orders se crean correctamente
3. [ ] Verificar que nuevos payments se crean correctamente
4. [ ] Monitorear tasa de errores
5. [ ] Verificar performance de DB

### Rollback Plan
Si algo falla:
```bash
# 1. Deter el backend actual
# 2. Restaurar backup pre-deploy
# 3. Revertir migraciones si es necesario
# 4. Redeploy versión anterior
```

## 📊 Comando de Verificación Todo en Uno

```bash
#!/bin/bash
echo "=== Verificación Pre-Deploy ==="

# 1. Compilación
echo "✓ Testing compilación..."
go build -o main ./cmd/main.go && echo "✓ Compilación OK" || echo "✗ Error de compilación"

# 2. Tests
echo "✓ Running tests..."
go test ./internal/services/payment/... -v && echo "✓ Tests OK" || echo "✗ Fallos en tests"

# 3. Imports
echo "✓ Verificando imports..."
if grep -q "preferenceMp" internal/services/payment/payment_service.go; then
  echo "✓ Preference import OK"
else
  echo "✗ Falta preference import"
fi

# 4. Archivos
echo "✓ Verificando archivos..."
required_files=(
  "migrations/002_add_delivery_and_payment_methods.sql"
  "CHECKOUT_INTEGRATION.md"
  "CHECKOUT_API_HELPERS.ts"
)

for file in "${required_files[@]}"; do
  if [ -f "$file" ]; then
    echo "✓ $file existe"
  else
    echo "✗ $file NO EXISTE"
  fi
done

echo "=== Verificación Completada ==="
```

## 🆘 Troubleshooting

### Error: "Cannot find package preferenceMp"
```bash
# Solución
go get github.com/mercadopago/sdk-go
```

### Error: "Invalid enum value"
- Verificar que los valores coincidan exactamente
- MercadoPago requiere mayúsculas: `MP_CARD` (no `mp_card`)

### Error: "Column delivery_method doesn't exist"
- La migración no fue aplicada
- Ejecutar: `mysql < migrations/002_add_delivery_and_payment_methods.sql`

### Error en Tests
- Verificar que mocks devuelven tipos correctos
- Verificar que interfaz MercadopagoClient es correcta

## 📞 Contacto y Soporte

Para preguntas:
1. Revisar `CHECKOUT_INTEGRATION.md`
2. Revisar `API_EXAMPLES.sh` para ejemplos
3. Revisar logs de ejecución
4. Verificar que todos los puntos del checklist estén completos

---

**Última Actualización:** 2026-05-08  
**Estado:** 🟢 Listo para Deploy

