# Cobertura de testes — 13 serviços Go (2026-09-22)

## Como medir (importante)

`go test -cover ./...` **engana** nesta base. Ele mede só o que o teste do
próprio pacote cobre, e vários pacotes são testados de fora — os handlers de
`in/http/<contexto>/` são exercitados pelos testes de integração do pacote pai
`in/http`, via `NewServer`.

Exemplo: `in/http/analysis` aparece como **0.0%** no `-cover` simples, mas os
handlers estão entre **73% e 90%** quando medidos com `-coverpkg`.

```bash
# cobertura da LÓGICA (o que a meta persegue)
go test -coverpkg=./internal/domain/...,./internal/application/... \
        -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -1

# o que ninguém exercita
go tool cover -func=cover.out | awk '$NF=="0.0%"'
```

## Meta acordada

**80% na lógica** (`domain/`, `application/`, mappers e DTOs com regra).

**Fora da conta**, por exigirem infraestrutura ou não terem comportamento:

| Fora da meta | Por quê |
|---|---|
| `cmd/*/main.go` | composition root |
| `server.go` `Start`/`Stop` | abre porta TCP de verdade |
| `rabbitmq/consumer.go` | precisa de broker |
| repositórios Mongo/Postgres | precisam de banco (sem testcontainers no repo) |
| `New*()` triviais | só atribuem campos |
| `dto/` sem lógica | testar seria testar o `encoding/json` |
| `port/in`, `port/out` | só declaração de interface |

**DTO com regra CONTA**: validação, `UnmarshalJSON` customizado ou campo
calculado precisam de teste como qualquer outra lógica.

## Situação por serviço

| Serviço | Antes | **Agora** | Meta 80% |
|---|---|---|---|
| srv-news-ingestion | 89.6% | **89.6%** | ✅ |
| bff-achadinhos | 88.0% | **88.0%** | ✅ |
| ms-analyst-finance | 87.4% | **87.4%** | ✅ |
| srv-sms-sender | 78.2% | **78.2%** | quase |
| srv-email-sender | 72.8% | **72.8%** | quase |
| ms-achadinhos | 71.9% | **72.3%** | quase |
| srv-data-collector | 71.8% | **71.8%** | quase |
| srv-audit | 71.0% | **71.0%** | quase |
| bff-auth | 62.7% | **67.5%** | falta |
| srv-llm-gateway | 56.4% | **62.3%** | falta |
| bff-core | 49.5% | **58.0%** | falta |
| bff-invest | 0.0% | 0.0% | n/a — ver nota |
| mock-sms-gateway | 0.0% | 0.0% | n/a — ver nota |

**bff-invest e mock-sms-gateway marcam 0% porque não têm lógica onde medir:**
o `bff-invest` tem `application/` só com `port/` e `domain/` vazio — toda a
regra está nos handlers (é o item da fase 8); o `mock-sms-gateway` é um mock.
A cobertura real do bff-invest, medida sobre todo o `internal/`, é **59.6%**.

## Bug encontrado escrevendo os testes

**Data race no circuit breaker** (`srv-news-ingestion` e `ms-analyst-finance`).
O `Manager` lia `b.state` **fora do mutex** para enviar à métrica, enquanto
`recordFailure`/`recordSuccess` escreviam esse campo sob lock. Detectado pelo
`-race` no primeiro teste de concorrência do pacote.

Correção: `allow`, `recordSuccess` e `recordFailure` devolvem o estado lido sob
o mesmo lock da escrita; o `Manager` usa o valor devolvido em vez de reler.

É o argumento a favor de testar `resilience/`: o pacote existe para o sistema
se comportar sob falha, e tinha uma corrida que só aparece sob carga.

## O que foi coberto nesta rodada

| Área | De | Para |
|---|---|---|
| `oauthsecret/crypto.go` | 0% | **90%** |
| `resilience/` (circuit + retry) | 0% | **57%** |
| `bff-core/domain/entities` | 0% | **79.7%** |
| `bff-core/domain/saga` | 40% | **53.2%** |
| `srv-llm-gateway` alertas e conversores | 0% | coberto |
| `bff-auth` validação de comando e parse de payload | 0% | coberto |

Os 10% que faltam em `crypto.go` são `if err != nil` de `NewCipher`/`NewGCM`,
inalcançáveis: a chave SHA-256 sempre tem 32 bytes válidos.

## Ainda descoberto (e por quê)

| Área | Situação |
|---|---|
| Repositórios Mongo (`runstore/`, `newsstore/`) | precisam de Mongo; exigiria adicionar testcontainers |
| Locks de janela (`scheduler_locks.go`, `newsstore/locks.go`) | idem — **é o maior risco remanescente**, porque são invariantes de concorrência |
| Publishers de auditoria | I/O puro, baixo risco funcional |
| `bff-invest` | sem camada de aplicação; cobrir exige a refatoração da fase 8 |

**Próximo passo com maior retorno:** adicionar testcontainers e cobrir os locks
de janela. Bug ali não aparece em desenvolvimento, só em produção com dois
jobs concorrendo.
