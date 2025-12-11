# Catálogo de Eventos

## Estructura Base

```go
type Event interface {
    EventID() string
    EventType() string
    AggregateID() string
    OccurredAt() time.Time
    Version() int
    Metadata() Metadata
}
```

## Eventos Implementados

### 1. PaymentRequested
**Cuándo:** Cliente crea pago  
**Publicado por:** CreatePaymentService  
**Consumido por:** PaymentOrchestrator  

```json
{
  "paymentId": "pmt_456",
  "userId": "user-123",
  "amount": 100.50,
  "currency": "ARS",
  "serviceId": "service-123"
}
```

### 2. WalletDebited
**Cuándo:** Se debita wallet  
**Publicado por:** PaymentOrchestrator  

```json
{
  "paymentId": "pmt_456",
  "userId": "user-123",
  "amount": 100.50,
  "previousBalance": 1000.00,
  "newBalance": 899.50
}
```

### 3. ExternalPaymentRequested
**Cuándo:** Solicita procesamiento externo  
**Publicado por:** PaymentOrchestrator  
**Consumido por:** ExternalGatewayMock  

### 4. ExternalPaymentSucceeded
**Cuándo:** Gateway confirma éxito  
**Publicado por:** ExternalGatewayMock  
**Consumido por:** PaymentOrchestrator  

```json
{
  "paymentId": "pmt_456",
  "externalTransactionId": "ext_tx_789"
}
```

### 5. ExternalPaymentFailed
**Cuándo:** Gateway rechaza  
**Publicado por:** ExternalGatewayMock  
**Consumido por:** PaymentOrchestrator  

```json
{
  "paymentId": "pmt_456",
  "reason": "CARD_DECLINED"
}
```

### 6. PaymentRefundRequested
**Cuándo:** Necesita refund  
**Publicado por:** PaymentOrchestrator  
**Consumido por:** PaymentOrchestrator (wallet handler)  

### 7. WalletCredited
**Cuándo:** Refund exitoso  
**Publicado por:** PaymentOrchestrator  

```json
{
  "paymentId": "pmt_456",
  "userId": "user-123",
  "amount": 100.50,
  "reason": "REFUND"
}
```

### 8. PaymentCompleted
**Cuándo:** Pago exitoso (final)  
**Publicado por:** PaymentOrchestrator  

### 9. PaymentFailed
**Cuándo:** Pago falla durante el procesamiento del orquestador  
**Publicado por:** PaymentOrchestrator  

```json
{
  "paymentId": "pmt_456",
  "reason": "WALLET_NOT_FOUND"
}
```

**Nota:** Si la validación falla en CreatePaymentService (wallet no existe, fondos insuficientes, currency mismatch), se retorna 400 Bad Request directamente sin crear el payment ni publicar eventos.

### 10. ExternalPaymentTimeout
**Definido pero no implementado en mock**

## Topología

**SNS Topic:** `payments-events`

**Suscripciones:**
- `payment-service-queue`: PaymentRequested, ExternalPayment*
- `wallet-service-queue`: PaymentRefundRequested
- `external-gateway-queue`: ExternalPaymentRequested

Cada cola tiene su DLQ con redrive policy de 3 intentos.

## Garantías

**At-Least-Once:** SQS puede entregar duplicados → idempotencia necesaria  
**No FIFO:** Orden no garantizado → validamos transiciones en DB  
**Event Sourcing:** Todos en EventStore antes de publicar
