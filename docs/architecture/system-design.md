# BFF-AUTH - System Design

## 1. Propósito e Domínio
- **Responsabilidade Principal:** Backend for Frontend de autenticação/autorização do KeepGuard: orquestra login, refresh, logout, validação de token, reset/change de senha, challenge de dispositivo, sessões, blacklist e lifecycle de conta, agregando chamadas aos microsserviços de domínio sem persistir estado próprio de negócio.
- **Domínio/Subdomínio:** IAM (Identity & Access Management) — superfície de borda para Auth, Device Trust, Session Management e Account Lifecycle.

## 2. Tech Stack Local
- **Linguagem & Framework:** Go 1.24 / Labstack Echo v4; config via Viper (`application.yml` + `application-{env}.yml`); validação com go-playground/validator; JWT com golang-jwt; HTTP outbound com Resty; Swagger (swaggo); logs Zap; métricas Prometheus (`promhttp` em porta dedicada).
- **Persistência e Cache:** Sem banco relacional/documental próprio. Redis (go-redis) para: presença de token de login (`tokenlogin:{codeUser}:{jwt}`), blacklist de dispositivo, rate limiting (Lua INCR/EXPIRE), cache de company/tenant, cache de usuário por code/email.
- **Mensageria:** RabbitMQ (wagslane/go-rabbitmq) como **publisher** (não consumidor):
  - Exchange topic de comunicação (default local: `ms-communication-exchange-local`, routing key `communication.message.send`) — reset/forgot password; fallback HTTP para `ms-communication` via circuit breaker.
  - Exchange de auditoria (default local: `srv-audit-exchange-local`, routing key `audit.event`) — eventos assíncronos do middleware de audit.

## 3. Arquitetura Interna
- **Padrão Utilizado:** Arquitetura Hexagonal (Ports & Adapters) com camadas Application (use cases) e Domain (ports/enums leves). Sem agregados DDD ricos — domínio concentrado em ports + enums de comunicação/template.
- **Módulos Principais:**
  - `cmd/bff-auth` — bootstrap, composition root, wiring de decorators e graceful shutdown.
  - `internal/adapters/inbound/http` — Echo server, handlers (`auth`, `message`), middlewares, mappers/DTOs de request/response.
  - `internal/adapters/outbound/http` — clients Resty (`auth`, `auth_user`, `company`, `communication`) + decorators (logging, metrics, retry, circuit breaker, Redis cache).
  - `internal/adapters/outbound/messaging` — publishers RabbitMQ (mensagens + audit) + decorators.
  - `internal/application` — use cases (`auth`, `message`) e ports de aplicação (`session`, `device`, `blacklist`, `lifecycle`); DTOs de comando.
  - `internal/application/port` — interfaces outbound (`AuthClient`, `UserClient`, `CompanyClient`, `CommunicationClient`).
  - `internal/domain/ports` — `messaging.MessagePublisher`, `audit.EventPublisher`.
  - `internal/infrastructure` — config, Redis, metrics, logger, resilience (gobreaker), validation, client IP.
  - `internal/pkg` — JWT claims/roles/authorities, erros padronizados.
  - `helm/` — deploy K8s (porta serviço 8381, metrics 9092).

## 4. Superfície de Contato (I/O)
- **Endpoints Expostos Principais:**
  - Infra: `GET /health`, `GET /swagger/*`, métricas em porta separada (default `9092`, path `/metrics`).
  - Auth (`/api/v1/auth`): `POST /login`, `/refresh`, `/logout`, `/validate`, `/change-password`, `/reset-password`, `/forgot-password`; device: `POST /device/challenge/send|verify`, `GET|POST /device/quick-revoke`.
  - Self-service (`/api/v1/users/me`, Bearer + token revocation): `POST /block`, `DELETE /`, sessões e blacklist de dispositivos.
  - Tenant/admin (Bearer + `session:read` / `session:write` ou roles `ADMIN`/`SYSTEM`): sessões/blacklist por `userId`, `GET /api/v1/sessions`, admin blacklist em `/api/v1/admin/devices/blacklist` e `/api/v1/devices/blacklist`.
- **Dependências Externas:**
  - **ms-auth** (HTTP, default `:8081`) — auth, sessões, blacklist, challenge, block/delete user; client principal com CB + retry + cache Redis.
  - **ms-company** (HTTP, default `:8083`) — resolve tenant → company (`GET /api/v1/companies/x-tenant-id/{tenantId}`); chamado em praticamente toda request via `CompanyResolveMiddleware`.
  - **ms-communication** (HTTP, default `:8082`) — `POST /api/v1/messages/send` como fallback do publisher RabbitMQ.
  - **RabbitMQ** — publicação de mensagens de comunicação e eventos de auditoria.
  - **Redis** — cache, rate limit e checagem de revogação de token/dispositivo.
  - Config declara `services.user` e `services.terms`, mas o client de usuário por e-mail usa a base URL do **ms-auth**; **terms não é consumido** no código de runtime atual.

## 5. Invariantes Locais e Observações
- **Multi-tenancy obrigatório na borda:** tenant vem do JWT (`tenant`) com prioridade sobre `X-Tenant-Id`; middleware resolve `companyId` e injeta no contexto; outbound propaga `X-Company-Id` e `X-Correlation-ID` / `X-Tenant-Id`.
- **Revogação na borda:** rotas autenticadas checam Redis `tokenlogin:{codeUser}:{jwt}`; ausência → `401 TOKEN_REVOKED`. Falha de Redis faz **bypass** (fail-open) com log de warn. Também bloqueia device na blacklist Redis.
- **Autorização de gestão de sessão:** `session:read` / `session:write` (authorities) ou roles `ADMIN`/`SYSTEM`.
- **Resiliência:** circuit breakers `ms-auth` e `ms-company` (abre com ≥10 requests e ≥70% falhas); decorators de retry com jitter; publisher de mensagem com CB e fallback HTTP para communication.
- **Rate limiting por rota** (Redis + regras em `rate_limit.rules` nos YAMLs de ambiente); login/forgot/challenge etc. com janelas configuráveis.
- **Só publica, não consome** filas; domínio de enums limitado a tipos de mensagem/comunicação/template.
- **Portas:** `application.yml` / local usam `8281`; default Viper, Swagger host e Helm usam `8381` — alinhar ambiente ao YAML ativo (`BFF_AUTH_ENV` / `APP_ENV`).
- **Prefixo de env:** `BFF_AUTH_*` (Viper com substitutor `.` → `_`). Sem `.env.example` no serviço; secrets Redis/RabbitMQ via YAML/env/Helm.
