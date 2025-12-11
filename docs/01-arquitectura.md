# Arquitectura del Sistema

## Visión General

Sistema de pagos event-driven con Go. La idea: separar responsabilidades con arquitectura hexagonal y que todo fluya mediante eventos.

## Componentes

```
Cliente
  ↓
HTTP API → CreatePaymentService → Publica eventos
                                        ↓
                                    SNS Topic
                                        ↓
                        ┌───────────────┼───────────────┐
                        ↓               ↓               ↓
                  payment-queue   wallet-queue   gateway-queue
                        ↓               ↓               ↓
              PaymentOrchestrator  (Refunds)   ExternalGatewayMock
```

**Tablas DynamoDB:**
- `Payments`: Estado de pagos
- `Wallets`: Balances de usuarios
- `EventStore`: Historial de eventos (event sourcing)
- `Idempotency`: Prevenir duplicados

**Colas SQS:**
- Cada una con su DLQ (3 reintentos)
- Visibility timeout: 30s

## Flujo Principal

**Happy Path:**
```
POST /payments
  → PaymentRequested
  → Valida wallet + debita
  → WalletDebited
  → ExternalPaymentRequested
  → Gateway procesa (mock: 200ms)
  → ExternalPaymentSucceeded
  → PaymentCompleted ✅
```

**Con Fallo del Gateway:**
```
POST /payments
  → PaymentRequested
  → Debita wallet
  → ExternalPaymentRequested
  → Gateway falla
  → ExternalPaymentFailed
  → PaymentRefundRequested
  → WalletCredited (compensación) ❌
```

## Capas

```
Domain (puro Go, sin deps)
  ↑
Application (use cases)
  ↑
Infrastructure (DynamoDB, SNS/SQS, HTTP)
```

**Domain:** Payment, Wallet, Events, Value Objects
**Application:** CreatePaymentService, PaymentOrchestrator
**Infrastructure:** Repositories, Publishers, Consumers, HTTP handlers

## Decisiones Clave

**¿Por qué monolito con colas?**
- Simple de deployar y desarrollar
- Las colas ya separan concerns para futura migración

**¿Por qué DynamoDB?**
- Compatible con LocalStack
- Event sourcing friendly
- Sin migrations

**¿Por qué SNS+SQS?**
- Desacoplamiento total
- Retry automático con DLQ
- Fan-out gratis

## Event Sourcing

Todos los eventos se guardan en `EventStore` antes de publicarse:
- Auditoría completa
- Replay posible
- Debugging temporal
