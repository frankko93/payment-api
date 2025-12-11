# 🚀 Cómo Usar la Payment API

## Requisitos

- Docker y Docker Compose
- Go 1.21+
- Make (opcional)

---

## 📦 Paso 1: Levantar LocalStack

> **Nota:** El Makefile ya configura automáticamente las variables para LocalStack.
> No necesitas crear archivos .env manualmente.

```bash
# Levantar LocalStack (DynamoDB, SNS, SQS)
docker-compose up -d

# Esperar 5 segundos para que esté listo
sleep 5
```

O usando Make:
```bash
make localstack-up
```

---

## 📊 Paso 2: Inicializar Base de Datos

**IMPORTANTE:** Debes crear las tablas ANTES de hacer el seed.

```bash
# Crear tablas DynamoDB y colas SQS
make init-db
```

Esto crea:
- Tablas: Payments, Wallets, Idempotency, EventStore
- Colas: payment-service-queue, wallet-service-queue, external-gateway-queue
- DLQs: payment-service-queue-dlq, wallet-service-queue-dlq, external-gateway-queue-dlq

---

## 🌱 Paso 3: Crear Datos de Prueba (Seed)

```bash
# Crear wallets de prueba
make seed
```

Esto crea:
- **user-123**: 1,000.00 ARS (fondos suficientes)
- **user-456**: 50.00 ARS (fondos insuficientes)

---

## ▶️ Paso 4: Levantar la API

Usando Make:
```bash
make run
```

La API estará en: **http://localhost:8080**

---

## 🧪 Paso 4: Probar la API

### Health Check

```bash
curl http://localhost:8080/health
```

Respuesta:
```
OK
```

---

### ✅ Crear Pago Exitoso (Fondos Suficientes)

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "amount": 100.50,
    "currency": "ARS",
    "serviceId": "service-123",
    "idempotencyKey": "test-key-001",
    "clientId": "web-app"
  }'
```

**Respuesta esperada:**
```json
{
  "paymentId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING"
}
```

El payment se procesará automáticamente via eventos.

---

### ❌ Crear Pago con Fondos Insuficientes

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-456",
    "amount": 100.00,
    "currency": "ARS",
    "serviceId": "service-456",
    "idempotencyKey": "test-key-002",
    "clientId": "web-app"
  }'
```

El payment se creará pero fallará por fondos insuficientes.

---

### 🔁 Probar Idempotencia

```bash
# Primer request
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "amount": 50.00,
    "currency": "ARS",
    "serviceId": "service-789",
    "idempotencyKey": "same-key-123",
    "clientId": "web-app"
  }'

# Segundo request (misma idempotencyKey)
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "amount": 50.00,
    "currency": "ARS",
    "serviceId": "service-789",
    "idempotencyKey": "same-key-123",
    "clientId": "web-app"
  }'
```

**Segunda respuesta:**
```json
{
  "paymentId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "ALREADY_PROCESSED"
}
```

Mismo paymentId, no se cobra dos veces ✅

---

## 🧪 Ejecutar Tests

### Tests Unitarios

```bash
go test ./tests/unit/... -v
```

O usando Make:
```bash
make test-unit
```

**Esperado:** 7/7 tests pasando ✅

---

### Tests de Integración

```bash
# Asegúrate de tener LocalStack corriendo
docker-compose up -d

# Ejecutar tests
RUN_INTEGRATION_TESTS=true go test ./tests/integration/... -v
```

O usando Make:
```bash
RUN_INTEGRATION_TESTS=true make test-integration
```

---

## 🛑 Detener Todo

```bash
# Detener LocalStack
docker-compose down
```

O usando Make:
```bash
make localstack-down
```

---

## 📋 Comandos Make Disponibles

```bash
make help              # Ver todos los comandos
make localstack-up     # Levantar LocalStack
make localstack-down   # Detener LocalStack
make seed              # Crear datos de prueba
make run               # Ejecutar la API
make test-unit         # Tests unitarios
make test-integration  # Tests de integración
make build             # Compilar binario
make clean             # Limpiar artifacts
```

---
### Payment con otra moneda

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "amount": 25.99,
    "currency": "USD",
    "serviceId": "service-usd",
    "idempotencyKey": "usd-payment-001",
    "clientId": "mobile-app"
  }'
```

### Payment con monto alto

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "amount": 999.99,
    "currency": "ARS",
    "serviceId": "premium-service",
    "idempotencyKey": "high-amount-001",
    "clientId": "web-app"
  }'
```

### Verificar errores de validación

```bash
# Sin userID (error)
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.00,
    "currency": "ARS",
    "serviceId": "service-123",
    "idempotencyKey": "error-test-001"
  }'
```

---