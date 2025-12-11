# Arquitectura del Sistema

## Visión General

Sistema de pagos event-driven con Go. La idea: separar responsabilidades con arquitectura hexagonal y que todo fluya mediante eventos.

## Componentes

```mermaid
flowchart TD
    Cliente([👤 Cliente])
    
    subgraph API["🚀 HTTP API"]
        CPS[CreatePaymentService]
    end
    
    subgraph Messaging["📨 Event Bus"]
        SNS[SNS Topic:<br/>payments-events]
    end
    
    subgraph Queues["📬 SQS Queues"]
        PQ[payment-queue]
        WQ[wallet-queue]
        GQ[gateway-queue]
    end
    
    subgraph Consumers["⚙️ Event Consumers"]
        PO[PaymentOrchestrator]
        WS[Wallet Refunds]
        GW[ExternalGatewayMock]
    end
    
    Cliente -->|POST /payments| CPS
    CPS -->|Publish events| SNS
    SNS --> PQ
    SNS --> WQ
    SNS --> GQ
    PQ --> PO
    WQ --> WS
    GQ --> GW
    
    style Cliente fill:#e3f2fd
    style API fill:#42a5f5,color:#fff
    style Messaging fill:#ff9800,color:#fff
    style PO fill:#4caf50,color:#fff
    style WS fill:#4caf50,color:#fff
    style GW fill:#9c27b0,color:#fff
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

```mermaid
sequenceDiagram
    actor Cliente
    participant API as POST /payments
    participant Bus as Event Bus
    
    Cliente->>API: Request payment
    API->>Bus: PaymentRequested
    Note over Bus: Valida wallet + debita
    Bus->>Bus: WalletDebited
    Bus->>Bus: ExternalPaymentRequested
    Note over Bus: Gateway procesa (mock: 200ms)
    Bus->>Bus: ExternalPaymentSucceeded
    Bus->>API: PaymentCompleted ✅
    API-->>Cliente: Payment successful
```

**Con Fallo del Gateway:**

```mermaid
sequenceDiagram
    actor Cliente
    participant API as POST /payments
    participant Bus as Event Bus
    participant Wallet as WalletService
    
    Cliente->>API: Request payment
    API->>Bus: PaymentRequested
    Bus->>Wallet: Debit wallet
    Wallet-->>Bus: Debited ✅
    Bus->>Bus: ExternalPaymentRequested
    Note over Bus: Gateway falla ❌
    Bus->>Bus: ExternalPaymentFailed
    Bus->>Bus: PaymentRefundRequested
    Bus->>Wallet: Credit wallet (compensación)
    Wallet-->>Bus: WalletCredited ✅
```

## Capas

```mermaid
graph BT
    subgraph Infrastructure["🔧 Infrastructure"]
        REPO[Repositories]
        PUB[Publishers]
        CONS[Consumers]
        HTTP[HTTP Handlers]
        DDB[(DynamoDB)]
        SNS[SNS/SQS]
    end
    
    subgraph Application["⚙️ Application (Use Cases)"]
        CPS[CreatePaymentService]
        PO[PaymentOrchestrator]
    end
    
    subgraph Domain["🎯 Domain (Pure Go, No Deps)"]
        PAY[Payment]
        WALL[Wallet]
        EVT[Events]
        VO[Value Objects]
    end
    
    Infrastructure --> Application
    Application --> Domain
    
    REPO --> DDB
    PUB --> SNS
    CONS --> SNS
    
    style Domain fill:#f3e5f5
    style Application fill:#fff3e0
    style Infrastructure fill:#e8f5e9
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
