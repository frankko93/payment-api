# Diagramas

## Flujo Happy Path

```
Cliente
  │
  └─▶ POST /payments
      │
      ▼
  CreatePaymentService
      │
      ├─▶ Check idempotency
      ├─▶ Save Payment (PENDING)
      └─▶ Publish PaymentRequested
          │
          ▼
      PaymentOrchestrator
          │
          ├─▶ Validate wallet
          ├─▶ Debit wallet
          ├─▶ Publish WalletDebited
          └─▶ Publish ExternalPaymentRequested
              │
              ▼
          ExternalGatewayMock
              │
              ├─▶ Simulate (200ms)
              └─▶ Publish ExternalPaymentSucceeded
                  │
                  ▼
              PaymentOrchestrator
                  │
                  ├─▶ Mark COMPLETED
                  └─▶ Publish PaymentCompleted ✅
```

## Flujo con Refund

```
Cliente
  │
  └─▶ POST /payments
      │
      ... (igual hasta gateway)
      │
      ▼
  ExternalGatewayMock
      │
      └─▶ Publish ExternalPaymentFailed ❌
          │
          ▼
      PaymentOrchestrator
          │
          ├─▶ Mark FAILED
          └─▶ Publish PaymentRefundRequested
              │
              ▼
          PaymentOrchestrator (wallet queue)
              │
              ├─▶ Credit wallet
              └─▶ Publish WalletCredited ✅
```

## Arquitectura de Capas

```
┌─────────────────────────┐
│   HTTP Handler          │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│   Application           │
│   • CreatePayment       │
│   • PaymentOrchestrator │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│   Domain                │
│   • Payment             │
│   • Wallet              │
│   • Events              │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│   Infrastructure        │
│   • DynamoDB            │
│   • SNS/SQS             │
└─────────────────────────┘
```

## Estados de Payment

```
        PENDING
           │
           ▼
       PROCESSING
       /        \
  Success    Failure
     │          │
     ▼          ▼
COMPLETED    FAILED
```

## SQS Retry Flow

```
Message → Consumer
            │
        ✅ Success → Delete
            │
        ❌ Error → Return to queue
            │
        Retry 1 (30s later)
            │
        ❌ Error
            │
        Retry 2 (30s later)
            │
        ❌ Error
            │
        Retry 3 (30s later)
            │
        ❌ Error
            │
        → DLQ (manual review)
```

## Event Sourcing

```
EventStore (DynamoDB)
  paymentId: pmt_123
  ├─ timestamp#1: PaymentRequested
  ├─ timestamp#2: WalletDebited
  ├─ timestamp#3: ExternalPaymentRequested
  ├─ timestamp#4: ExternalPaymentSucceeded
  └─ timestamp#5: PaymentCompleted

Query(paymentId) → Replay → Reconstruct Payment
```

## Componentes

```
┌──────────────────┐
│   Payment API    │
│  (single process)│
└────────┬─────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
DynamoDB   SNS/SQS
(4 tables) (1 topic, 3 queues)
```
