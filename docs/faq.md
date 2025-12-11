# FAQ

## ¿Por qué monolito con eventos?

Arquitectura event-driven dentro de un monolito:
- Simplicidad de deployment (un solo proceso)
- Fácil migrar a microservicios (las colas ya separan concerns)

## ¿Cómo funciona la idempotencia?

Cada request tiene un `idempotencyKey` único:

```
idempotencyKey → {paymentId, status, response}
```

Si llega duplicado: retornamos el resultado anterior sin crear payment nuevo.

## ¿Qué eventos se emiten?

**Pago exitoso:**
1. PaymentRequested
2. WalletDebited
3. ExternalPaymentRequested
4. ExternalPaymentSucceeded
5. PaymentCompleted

**Pago fallido con refund:**
1. PaymentRequested
2. WalletDebited
3. ExternalPaymentRequested
4. ExternalPaymentFailed
5. PaymentRefundRequested
6. WalletCredited

## ¿Los eventos pueden llegar fuera de orden?

Sí, SQS Standard no garantiza FIFO. Mitigamos validando transiciones de estado en DB.

## ¿Qué hago con mensajes en DLQ?

1. Revisar mensaje y logs
2. Si fue bug → fixear y republicar
3. Si fue outage → republicar cuando esté OK

```bash
aws sqs receive-message \
  --queue-url http://localhost:4566/.../payment-service-queue-dlq
```
