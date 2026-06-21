# syntax=docker/dockerfile:1
FROM golang:1.22 AS build

WORKDIR /src

# Copia go.mod e go.sum primeiro para aproveitar cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bff-auth ./cmd/bff-auth \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/healthcheck ./deploy/healthcheck

# Stage final - imagem mínima
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copia o binário principal
COPY --from=build /out/bff-auth /app/bff-auth

# Copia o utilitário de healthcheck
COPY --from=build /out/healthcheck /app/healthcheck

# Copia arquivos de configuração
COPY --from=build /src/application*.yml /app/

# Usuário não-root para segurança
USER 65532:65532

# Porta da aplicação
EXPOSE 8381

# Configura URL padrão do healthcheck
ENV HEALTHCHECK_URL=http://localhost:8381/health

# Comando de entrada
ENTRYPOINT ["/app/bff-auth"]
