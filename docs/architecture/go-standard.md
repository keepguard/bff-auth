# Padrão de Estrutura e Nomenclatura — Serviços Go do KeepGuard

Documento normativo para as **13 aplicações Go** do monorepo (`bff-*`, `ms-*`, `srv-*`,
`mock-*`), nos domínios `achadinhos/`, `investbot/` e `keepguard-core/`.

**Referência canônica:** `keepguard-core/backend/bff/bff-auth`. Em caso de dúvida não
coberta aqui, o que o `bff-auth` faz é o padrão.

> Serviços Java (`ms-*` do keepguard-core) e Python (`srv-token-manager`) têm padrão
> próprio e estão **fora** do escopo deste documento. Para Java, a referência é `ms-auth`.

---

## 1. Versão do Go

| Item | Valor |
|---|---|
| Versão alvo | **1.27.1** |
| Diretiva `go` no `go.mod` | `go 1.27.1` (patch completo) |
| Imagem base de build | `golang:1.27` ou `golang:1.27-alpine` |

**Regra:** a diretiva `go` do `go.mod` e a imagem do Dockerfile devem estar sempre na
mesma linha de release. Divergência aqui **quebra o deploy** — já aconteceu: com
`go.mod` em 1.27.1 e imagem `golang:1.24-alpine`, o build falha com
`go.mod requires go >= 1.27.1 (running go 1.24.13; GOTOOLCHAIN=local)`.

**Sobre `toolchain`:** não adicione a diretiva `toolchain go1.27.1`. O `go mod tidy` a
remove automaticamente quando `go` já traz o patch completo, por ser redundante.

Ao subir de versão, atualizar sempre **os dois** (`go.mod` + Dockerfile), rodar
`go mod tidy` e validar com `go build ./... && go vet ./... && go test ./...`.

---

## 2. Layout de diretórios

```
<serviço>/
├── cmd/<serviço>/main.go             # único entrypoint; nome = nome do serviço
├── deploy/
│   ├── Dockerfile                    # multi-stage (ver regra 5)
│   └── healthcheck/main.go           # binário de healthcheck da imagem
├── db/
│   ├── migrations/                   # só se o serviço tem banco
│   └── seed/
├── docs/architecture/system-design.md
├── helm/templates/
├── application.yml                   # base
├── application-{dev,local,prod}.yml  # os três são obrigatórios (ver nota)
├── script-deploy-github-<serviço>.sh
├── script-deploy-k8s-prod.sh
├── .gitignore                        # ver regra 6
└── internal/
    ├── adapters/                     # tudo que fala com o mundo externo
    │   ├── in/                       # quem CHAMA o serviço
    │   │   ├── http/
    │   │   │   ├── server.go         # montagem do servidor e rotas
    │   │   │   ├── interfaces.go     # interfaces consumidas pelos handlers
    │   │   │   ├── dto/              # request/response HTTP — NUNCA no handler
    │   │   │   ├── handlers/         # um arquivo por recurso
    │   │   │   ├── mapper/           # DTO <-> domínio
    │   │   │   └── middleware/
    │   │   ├── scheduler/            # cron/jobs
    │   │   └── rabbitmq/             # consumers
    │   └── out/                      # quem o serviço CHAMA
    │       ├── http/
    │       │   ├── client/           # um arquivo por serviço externo
    │       │   ├── decorator/<domínio>/  # retry, cache, circuit breaker, log, métrica
    │       │   ├── dto/
    │       │   └── mapper/
    │       ├── messaging/{audit,decorator,rabbitmq}/
    │       └── postgres/             # repositórios
    ├── application/                  # casos de uso — orquestra, não decide regra
    │   ├── <contexto>/
    │   │   ├── <ação>_usecase.go
    │   │   └── interfaces.go
    │   ├── dto/                      # commands e resultados internos
    │   └── port/                     # portas de saída consumidas pela aplicação
    ├── domain/                       # regra de negócio pura, sem dependência externa
    │   ├── entities/
    │   ├── enums/
    │   ├── valueobjects/
    │   └── ports/<subdomínio>/
    ├── infrastructure/               # técnico, sem regra de negócio
    │   ├── config/                   # obrigatório
    │   ├── logger/                   # obrigatório
    │   ├── metrics/                  # obrigatório
    │   ├── validation/               # obrigatório onde há entrada externa
    │   ├── resilience/               # obrigatório onde há chamada externa
    │   ├── cache/                    # quando usa Redis
    │   ├── swagger/                  # quando expõe API documentada
    │   ├── clientip/ requestmeta/    # quando aplicável
    └── pkg/                          # utilitários transversais do serviço
```

### Nota sobre `application-prod.yml`

O carregamento é `application.yml` (base) + merge de `application-<APP_ENV>.yml`, com o
erro do merge ignorado (`_ = viper.MergeInConfig()`) — ausência do arquivo de perfil não
quebra o boot, só não sobrepõe nada.

Faltam em **`srv-email-sender`** e **`mock-sms-gateway`**. Em ambos o `helm/values.yaml`
traz `appEnv: dev`, ou seja, rodam com perfil `dev` **em produção** e um
`application-prod.yml` sequer seria lido.

Esses dois arquivos **não foram criados** na fase de higiene de propósito: o conteúdo
correto (SMTP, credenciais, hosts) é configuração real de produção, não há como inferir,
e um arquivo adivinhado é pior que nenhum. Criar exige decidir junto se o `appEnv` desses
serviços passa a `prod` — mudança de runtime, não de estrutura.

### Nomes de diretório proibidos

| Proibido | Correto |
|---|---|
| `internal/adapters/inbound/` | `internal/adapters/in/` |
| `internal/adapters/outbound/` | `internal/adapters/out/` |
| `internal/core/domain/` | `internal/domain/entities/` |
| `internal/core/ports/` | `internal/application/port/{in,out}/` |
| `internal/core/service/` | `internal/application/<contexto>/` |
| `internal/domain/ports/` | `internal/application/port/{in,out}/` |
| `internal/{auth,client,tools}/` na raiz | dentro da camada hexagonal correspondente |

---

## 3. Nomenclatura de arquivos

**Princípio: o nome do arquivo diz a que pertence, sem precisar abrir.**

Foi por violar isso que o padrão virou problema: um `usecase.go` dentro de
`application/curadoria/` não revela qual caso de uso implementa, e dois arquivos
`ports.go` em serviços diferentes não dizem nada sobre o que declaram.

| Proibido | Obrigatório | Exemplo |
|---|---|---|
| `usecase.go`, `usecases.go` | `<ação>_usecase.go` | `login_usecase.go`, `aprovar_candidato_usecase.go` |
| `ports.go` com muitas interfaces soltas | um arquivo por tema | `product_repository.go`, `scraper.go`, `notifier.go` |
| `handler.go`, `handlers.go` | `<recurso>_handlers.go` | `auth_handlers.go`, `product_handlers.go` |
| `dto.go`, `types.go`, `models.go` | `<entidade>_{request,response,command}.go` | `login_command.go`, `auth_response.go` |
| `client.go` | `<domínio>_client.go` | `auth_client.go`, `company_client.go` |
| `repository.go` | `<entidade>_repository.go` | `product_repository.go` |
| `service.go` | `<contexto>_service.go` ou vira `_usecase.go` | |
| `mapper.go` genérico | `<entidade>_mapper.go` | `message_mapper.go` |

### Portas: `application/port/{in,out}`

Quem declara porta é a **aplicação**, não o domínio:

- `port/in/` (package `inport`) — **driving**: o que o mundo pede ao app.
  Implementado pelos use cases, chamado pelos handlers.
- `port/out/` (package `outport`) — **driven**: o que o app pede ao mundo.
  Implementado pelos adapters outbound, chamado pelos use cases.

`domain/ports/` está **proibido em serviço novo**: porta é contrato da aplicação
com o mundo externo, não regra de negócio. O domínio tem que poder ser lido sem
saber que existe HTTP ou Postgres.

Migração concluída nos 13 serviços (2026-09-22). `port/in` existe só onde as
portas de entrada foram declaradas explicitamente (`ms-analyst-finance`,
`srv-news-ingestion`); nos demais a interface do caso de uso faz esse papel.

Um arquivo por tema — nunca um `ports.go` com 20 interfaces soltas.

### Entrada nunca chama saída

`in/` não importa `out/`. Handler que fala com cliente HTTP ou repositório
direto pulou a aplicação inteira. Caminho: `in/ → port/in → use case → port/out → out/`.

Verificação: `grep -rn "adapters/out" internal/adapters/in --include="*.go"`
tem que voltar vazio.

Vale para **todo** adapter de entrada — handler HTTP, scheduler e consumer de
fila. O scheduler é o que mais escapa por parecer "interno".

### O handler depende da porta, não do use case

Receber `*analyze.UseCase` (struct concreto) no handler fura a inversão de
dependência. O adapter conhece **só a interface** de `application/port/in`.

Consequência: o adapter não importa o pacote do use case para nada. Erros
sentinela, commands e helpers de contexto moram em `application/dto`; quem
constrói caso de uso é o composition root.

Interfaces segregadas por contexto (ISP), não uma porta com 31 métodos.

### Organização de `in/http`: por contexto, a partir de 3

Com 3 ou mais contextos, cada um vira fatia vertical completa:

```
in/http/
├── analysis/{dto,mapper}/ + handler.go + analysis_handlers.go
├── catalog/{dto,mapper}/  + handler.go + catalog_handlers.go
├── watchlist/{dto,mapper}/+ handler.go + watchlist_handlers.go
├── httperr/               helpers compartilhados
└── server.go
```

Com 1 ou 2 contextos, fica a forma simples `in/http/{dto,mapper,handlers}/`.

**O `dto/` não vai dentro de `handlers/`** — isso faria o handler virar dono do
DTO. O contexto vem antes da divisão técnica; `dto/` e `mapper/` são irmãos.

**Um `Handler` por contexto**, recebendo só as portas que usa (Go não permite
métodos do mesmo struct em pacotes diferentes, e aqui isso ajuda).

Em `application/`, o `dto` fica **neutro** (`application/dto`): `port/in`
importa os commands e o pacote do use case implementa `port/in` — o dto dentro
do contexto criaria ciclo.

Aplicado em `ms-analyst-finance` (2026-09-22). Revelou um `mapper.go` de 569
linhas com três contextos misturados e 4 handlers no contexto errado.

### O `out/` segue as mesmas regras

Primeiro a tecnologia, depois o contexto: `out/http/<serviço>/`,
`out/messaging/<tema>/`, `out/mongo|postgres/<contexto>/`, `out/redis/`.

O adapter de saída tem `dto/` e `mapper/` próprios — o JSON do parceiro não
chega ao domínio. Em banco, a struct com tags `bson`/`db` é um **document**,
separado do repositório e do mapper.

**A interface do cliente não fica no adapter**: ela é a porta, em
`application/port/out`. Exceção: interface que é dependência interna do adapter
e não atravessa camada (ex.: `TokenSource`).

Nomes: `<serviço>_client.go`, `<tema>_publisher.go`, `<entidade>_repository.go`,
`<entidade>_document.go`. Proibidos: `client.go`, `adapter.go`, `mongo.go`.

### Onde fica o mapper (e por que não dentro do dto)

`dto/` e `mapper/` são **pastas irmãs**, cada uma seu pacote:

```
adapters/inbound/http/
├── dto/      → struct + tag JSON. Dado puro, NAO importa domain
├── mapper/   → traduz. Importa dto E domain
└── handlers/ → importa dto e mapper
```

A regra é a direção da dependência: se a função de conversão mora dentro de `dto/`,
o DTO passa a importar o domínio e deixa de ser dado puro — o adaptador contamina a
borda que existia justamente para isolar. O mapper é quem pode conhecer os dois lados.

Vale nos dois sentidos, como no `bff-auth`:
`inbound/http/mapper/` (HTTP → aplicação) e `outbound/http/mapper/` (aplicação → serviço externo).

### Todo caso de uso termina em `_usecase.go`

Se o arquivo declara ou implementa métodos de um `UseCase`/`service` da camada
`application`, o nome termina em `_usecase.go` — **sem exceção**. Um `compare.go` com
`CompareAssetsUseCase.Execute` dentro obriga a abrir o arquivo para descobrir o que é.

Vale também quando o caso de uso está partido em vários arquivos por assunto: cada
parte leva o sufixo (`catalog_usecase.go`, `proactive_usecase.go`, `batch_usecase.go`).

Continuam sem sufixo só os arquivos que **declaram a interface**, não a implementam:
`interfaces.go` e `port.go` dentro do pacote de contexto (padrão do `bff-auth`).

### O nome do pacote já conta como contexto

Nome de arquivo Go é lido como `pacote/arquivo`. Quando o pacote já delimita o assunto,
repetir o nome dele no arquivo só faz ruído — `alert/alert_port.go` não diz nada que
`alert/port.go` já não dissesse.

**Ficam como estão** (é o padrão do `bff-auth`):

- `port.go` e `interfaces.go` dentro de um pacote de contexto — `application/auth/interfaces.go`,
  `application/blacklist/port.go`
- `client.go`, `mapper.go` dentro de um pacote que já nomeia o destino —
  `outbound/mercadolivre/client.go`

**Precisam de nome melhor** quando o pacote NÃO delimita:

- `usecase.go` — o pacote diz o contexto (`curadoria`), mas não a ação. Quebrar por ação:
  `descobrir_usecase.go`, `buscas_usecase.go`, `candidatos_usecase.go`.
- `ports.go` em qualquer nível — o pacote é genérico e o arquivo acumula
  dezenas de interfaces de assuntos diferentes. Quebrar por tema.
- `handler.go`, `dto.go`, `types.go`, `models.go` num pacote genérico como `http/`.

### Testes

Arquivo de teste acompanha o arquivo testado: `login_usecase.go` → `login_usecase_test.go`.

---

## 4. Nomenclatura de tipos

Padrão do `bff-auth` — **interface pública + implementação privada**:

```go
// interfaces.go
type LoginUseCase interface {
    Execute(ctx context.Context, cmd dto.LoginCommand) (*dto.LoginResult, error)
}

// login_usecase.go
type loginUseCaseImpl struct { ... }

func NewLoginUseCase(authClient authclient.AuthClient) LoginUseCase {
    return &loginUseCaseImpl{authClient: authClient}
}
```

| Proibido | Obrigatório |
|---|---|
| `type useCase struct` (genérico, minúsculo) | `type loginUseCaseImpl struct` |
| `type UseCase struct` exportado sem interface | interface `LoginUseCase` + impl privada |
| Struct-Deus com 20+ métodos públicos | um caso de uso por responsabilidade |

**Construtor:** sempre `New<Nome>` retornando a **interface**, nunca o struct concreto.

### DTOs nunca moram em handler

Todo struct de request/response vai em `adapters/inbound/http/dto/`. Um handler
declara **apenas** o próprio struct receptor (`type AuthHandlers struct`) e seus métodos.

---

## 5. Build e Docker

**Regra: todo serviço compila DENTRO do Docker, em multi-stage.**

Hoje 6 serviços compilam no Mac do desenvolvedor e o Dockerfile só faz `COPY` do
binário. Isso torna o build **não reproduzível** (depende da versão de Go instalada na
máquina de quem deployou) e faz a diretiva `go` do `go.mod` mentir sobre o que roda em
produção.

Modelo (do `bff-auth`):

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/<serviço> ./cmd/<serviço> \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/healthcheck ./deploy/healthcheck

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/<serviço> /app/<serviço>
COPY --from=build /out/healthcheck /app/healthcheck
COPY --from=build /src/application*.yml /app/
USER 65532:65532
EXPOSE <porta>
ENTRYPOINT ["/app/<serviço>"]
```

- Localização: `deploy/Dockerfile` (não na raiz).
- Imagem final: `gcr.io/distroless/base-debian12` (preferida) ou `alpine:3.20`.
- Usuário não-root obrigatório.
- Todo build para a VPS Hostinger usa `--platform linux/amd64`.

### Healthcheck: duas abordagens válidas

| Abordagem | Quando usar | Quem usa hoje |
|---|---|---|
| **Binário** `deploy/healthcheck/main.go` | Imagem **distroless** (não tem shell nem `curl`) | bff-auth, bff-core, bff-achadinhos, ms-achadinhos |
| `curl` no `HEALTHCHECK` | Imagem **alpine** (traz `curl` via `apk add`) | os outros 8 |

Ambas atendem. A do binário é preferível por permitir distroless (superfície de ataque
menor e imagem menor), e é para onde convergir ao migrar um serviço para distroless.
**Não** criar `deploy/healthcheck/` num serviço alpine que já usa `curl` e funciona — é
troca de runtime sem ganho imediato.

> Exceção a corrigir: `mock-sms-gateway` não tem `HEALTHCHECK` de nenhum tipo.

> **Nota operacional:** o build amd64 sob emulação em Mac ARM é lento (>10 min) e se
> mostrou instável no Colima — falhas intermitentes de TLS e crash de runtime que não
> são do código. Use `run_in_background` e acompanhe o log.

---

## 6. `.gitignore` e artefatos

**Binário compilado nunca é versionado.** Hoje há 3 binários rastreados somando ~32 MB
(`bff-auth/.bin/healthcheck`, `bff-core/.bin/healthcheck`,
`srv-sms-sender/srv-sms-sender-linux-amd64`) — todos são regenerados pelo script de
deploy ou pelo Dockerfile, nenhum é consumido por nada.

> Acrescentar ao `.gitignore` **não** destrackeia arquivo já versionado. É preciso
> `git rm --cached <arquivo>` explicitamente.

`.gitignore` padrão (substituir `__SERVICE__` pelo nome do serviço):

```gitignore
# --- Binários compilados (NUNCA versionar) ---
*.exe
*.exe~
*.dll
*.so
*.dylib
/__SERVICE__
/__SERVICE__-linux-amd64
.bin/
bin/

# --- Artefatos de teste e cobertura ---
*.test
*.out
coverage*.out

# --- Dependências e workspace ---
vendor/
go.work
go.work.sum

# --- IDE ---
.idea/
*.iml
.vscode/
*.swp
*.swo
*~

# --- Sistema operacional ---
.DS_Store
Thumbs.db

# --- Logs ---
*.log

# --- Ambiente e segredos ---
.env
.env.local
.env.*.local
secure/
*.pem
credentials.json
token.json

# --- Backups do script de deploy ---
.deploy-backup/

# --- Documentação gerada (swagger) ---
docs/swagger.json
docs/swagger.yaml
docs/docs.go
```

### Swagger: `docs.go` gerado

Quem usa `swag` importa o pacote `docs` no `main.go`. Para o build funcionar a partir de
um clone limpo, o pacote precisa existir mesmo sem rodar o `swag`. O padrão é o do
`bff-core`: versionar um **`docs/swagger.go` placeholder** (`package docs`) e ignorar o
`docs/docs.go` gerado.

> Divergência conhecida: o `bff-auth` versiona o `docs.go` gerado. Convergir para o
> modelo do `bff-core` na fase de higiene.

---

## 7. Checklist de conformidade

Ao criar ou revisar um serviço Go:

- [ ] `go.mod` em `go 1.27.1`, sem diretiva `toolchain`
- [ ] Dockerfile em `golang:1.27*`, multi-stage, em `deploy/Dockerfile`
- [ ] `cmd/<serviço>/main.go` como único entrypoint
- [ ] Healthcheck presente — `deploy/healthcheck/main.go` (obrigatório em imagem distroless) ou `curl` no `HEALTHCHECK`
- [ ] `application{,-dev,-local,-prod}.yml` — os quatro
- [ ] `internal/{adapters,application,domain,infrastructure}` — sem `in/`, `out/`, `core/`
- [ ] Portas em `application/port/{in,out}` — sem `domain/ports` em serviço novo
- [ ] `inbound/http/` com `dto/`, `mapper/`, `handlers/` como pacotes separados
- [ ] Nenhuma função de conversão dentro de `dto/`
- [ ] Todo arquivo que implementa caso de uso termina em `_usecase.go`
- [ ] `infrastructure/`: `config`, `logger`, `metrics` sempre; `validation` e `resilience` quando aplicável
- [ ] Nenhum arquivo `usecase.go`, `ports.go`, `handler.go`, `dto.go`, `types.go`, `models.go`
- [ ] Nenhum struct de request/response declarado fora de `dto/`
- [ ] Casos de uso: interface exportada + impl privada + `New<Nome>` retornando interface
- [ ] `.gitignore` conforme regra 6; nenhum binário rastreado
- [ ] `go build ./... && go vet ./... && go test ./...` verdes

---

## 8. Estado atual (2026-09-22)

Aderência ao padrão após a atualização para Go 1.27.1:

| Projeto | Camada | Aderência | Principais desvios |
|---|---|---|---|
| **bff-auth** | core/bff | 🟢 referência | versiona `docs.go` gerado |
| bff-core | core/bff | 🟢 alta | — |
| ms-achadinhos | achadinhos/ms | 🟡 média | 8 `usecase.go`, 6 `port.go`, `ports.go` (424 linhas), `dto.go`, nomes PT/EN |
| bff-achadinhos | achadinhos/bff | 🟡 média | sem `application/port/{in,out}` |
| srv-data-collector | core/srv | 🟡 média | handlers sem subpasta `handlers/`, 3 `port.go`, build no host |
| srv-news-ingestion | investbot/srv | 🟢 alta | segue o padrão |
| srv-audit | core/srv | 🟡 média | sem `application/port`, `mapper/`, `middleware/`; build no host |
| srv-llm-gateway | core/srv | 🟠 baixa | `handlers.go` (317), `dto.go` (246), `ports.go`; build no host |
| srv-email-sender | core/srv | 🟠 baixa | `ports.go` monolítico; sem `logger`/`validation`/`resilience`; build no host |
| srv-sms-sender | core/srv | 🟠 baixa | idem; binário versionado; build no host |
| **ms-analyst-finance** | investbot/ms | 🔴 crítica | `handler.go` 2.036 linhas com ~40 DTOs inline; `usecase.go` 1.142 linhas |
| **bff-invest** | investbot/bff | 🔴 crítica | `handler.go` 707 linhas; **sem camada `application/`** |
| **mock-sms-gateway** | core/mock | 🔴 crítica | `in/out/core`; `handler.go`/`repository.go`/`models.go`/`ports.go`; zero testes |

---

## 9. Fases de padronização

| Fase | Escopo | Risco | Status |
|---|---|---|---|
| 1 | Go 1.27.1 em todos + correção dos Dockerfiles divergentes | Baixo | ✅ **concluída** |
| 0 | Este documento + `.gitignore` padrão | Nulo | ✅ **concluída** |
| 2 | Higiene: destrackear binários (~30 MB) + `cmd/` dos achadinhos | Baixo | ✅ **concluída** |
| 3 | Renomear arquivos genéricos (`usecase.go` → `<ação>_usecase.go` etc.) | Baixo | pendente |
| 4 | Extrair DTOs de handlers; quebrar handlers monolíticos | Médio | pendente |
| 5 | Migrar `domain/ports` → `application/port/out` (9 serviços) | Médio | ✅ **concluída** |
| 5b | `in/out`/`core/` do mock-sms-gateway → inbound/outbound/domain | Médio | pendente |
| 6 | Completar `infrastructure/` (logger, validation, resilience) | Médio | pendente |
| 7 | Migrar os 6 serviços que compilam no host para build multi-stage | Médio | pendente |
| 8 | Refatoração estrutural: `application/` no bff-invest; quebrar o UseCase de 1.142 linhas | **Alto** | pendente |

**Regras de execução:**

- Cada serviço é um **repositório Git independente** → um commit e um deploy por serviço.
- Antes de qualquer fase: `git status` no repo e baseline de `build`/`vet`/`test`.
- Nenhuma fase até a 7 altera contrato de API — por isso **não** atualizam
  `system-design.md` (ver regra 1 do `CLAUDE.md` raiz). A fase 8 pode alterar, e aí a
  atualização é obrigatória.
- Validação obrigatória por serviço antes do deploy:
  `go build ./... && go vet ./... && go test ./...` verdes, comparados ao baseline.

---

## 10. Fora de escopo

- Serviços **Java** (`ms-auth`, `ms-user`, `ms-billing`, `ms-company`, `ms-knowledge`,
  `ms-communication`, `ms-user-consents`, `ms-ai-guardian`) — referência: `ms-auth`.
- Serviço **Python** (`srv-token-manager`).
- **Frontends** (`achadinhos/front`, `keepguard-core/frontend/backoffice`).
- **Tradução PT→EN** de identificadores do `ms-achadinhos` (`curadoria`, `rotacao`) —
  decisão de produto, tratar em separado.
- **Aumento de cobertura de testes** — as fases mantêm verde o que já existe, sem
  escrever testes novos.
