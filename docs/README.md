# Documentación de Arquitectura

Documentación técnica del sistema de pagos event-driven.

## Documentos

- **[Arquitectura](./01-arquitectura.md)** - Componentes y flujos principales
- **[Eventos](./02-eventos.md)** - Catálogo de eventos
- **[Manejo de Errores](./03-manejo-errores.md)** - Retry, DLQ y compensación
- **[Diagramas](./diagramas.md)** - Flujos visuales
- **[FAQ](./faq.md)** - Preguntas frecuentes

## Quick Reference

**Flujo Happy Path:**

```mermaid
flowchart LR
    A([POST /payments]) --> V{Validate<br/>Wallet}
    V -->|OK| B[PaymentRequested]
    V -->|Fail| E[400 Error]
    B --> C[Debit Wallet]
    C --> D[External Gateway]
    D --> F[PaymentCompleted ✅]
    
    style F fill:#4caf50,color:#fff
    style E fill:#f44336,color:#fff
```

**Flujo con Fallo:**

```mermaid
flowchart LR
    A([POST /payments]) --> B[Debit Wallet]
    B --> C[Gateway Fails ❌]
    C --> D[Refund]
    D --> E[WalletCredited]
    
    style C fill:#f44336,color:#fff
    style E fill:#ff9800,color:#fff
```

**Estados de Payment:**

```mermaid
stateDiagram-v2
    [*] --> PENDING
    PENDING --> PROCESSING
    PROCESSING --> COMPLETED
    PROCESSING --> FAILED
    COMPLETED --> [*]: ✅
    FAILED --> [*]: ❌
```

**Garantías:**
- ✅ Idempotencia (safe to retry)
- ✅ Compensación automática (Saga)
- ✅ Event sourcing (auditoría completa)
- ✅ At-least-once delivery
- ❌ FIFO ordering (mitigado con estado en DB)

## Stack

- **Go** 1.21+
- **DynamoDB** (4 tablas)
- **SNS** (1 topic)
- **SQS** (3 queues + 3 DLQs)
- **LocalStack** (desarrollo)
