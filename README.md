# URL Shortener API

API REST em Go para criar links curtos, redirecionar acessos e consultar métricas básicas.

## Endpoints

| Método | Rota | Descrição |
| --- | --- | --- |
| `POST` | `/api/v1/urls` | Cria uma URL curta. |
| `GET` | `/{code}` | Redireciona para a URL original. |
| `GET` | `/api/v1/urls/{code}` | Retorna dados e total de acessos. |
| `GET` | `/health` | Verifica API e banco configurado. |
| `GET` | `/metrics` | Métricas no formato Prometheus. |

## Exemplo

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","expires_at":"2030-01-01T00:00:00Z"}'
```

`expires_at` é opcional e deve estar no futuro, em RFC 3339. Após expirar, o redirecionamento responde `410 Gone` e não contabiliza acessos.

## Executar localmente

Pré-requisito: Go 1.23+.

```bash
go run ./cmd/api
go test ./...
```

Sem `DATABASE_URL`, a API usa armazenamento em memória — adequado apenas para desenvolvimento e testes.

## Docker e PostgreSQL

```bash
docker compose up --build
```

O serviço `migrate` aplica os arquivos versionados em `migrations/` antes da API iniciar. Para apagar o banco local e iniciar de novo: `docker compose down -v`.

## Configuração

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `PORT` | `8080` | Porta HTTP da API. |
| `DATABASE_URL` | vazia | String de conexão PostgreSQL; habilita persistência. |
| `RATE_LIMIT_PER_MINUTE` | `60` | Máximo de requisições por IP por minuto; `0` desabilita. |

## Recursos de produção

- PostgreSQL com migrations SQL versionadas.
- Rate limit em memória por IP, com resposta `429` e `Retry-After`.
- Logs estruturados, `/health` e `/metrics`.
- CI em GitHub Actions: `go vet`, testes com detector de corrida e build da imagem Docker.
