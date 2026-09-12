# URL Shortener API

API REST em Go para criar links curtos, redirecionar acessos e consultar métricas básicas.

> Projeto construído como parte de uma trilha pública de estudos focada em backend, APIs, testes e práticas de engenharia.

## Escopo da primeira versão

- Criar uma URL curta a partir de uma URL de destino.
- Redirecionar acessos por meio do código curto.
- Consultar a quantidade de acessos de cada link.
- Validar URLs e impedir códigos duplicados.
- Cobrir os fluxos principais com testes automatizados.

Nesta primeira etapa os dados ficam em memória para manter o ciclo de desenvolvimento rápido. PostgreSQL, Docker e persistência entram na próxima etapa, sem alterar o contrato HTTP.

## Endpoints

| Método | Rota | Descrição |
| --- | --- | --- |
| `POST` | `/api/v1/urls` | Cria uma URL curta |
| `GET` | `/{code}` | Redireciona para a URL original |
| `GET` | `/api/v1/urls/{code}` | Retorna dados e total de acessos |
| `GET` | `/health` | Verifica a saúde da API |

## Exemplo

Crie um link:

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev"}'
```

Resposta esperada:

```json
{
  "code": "a1b2c3",
  "short_url": "http://localhost:8080/a1b2c3",
  "original_url": "https://go.dev",
  "clicks": 0
}
```

## Executar localmente

Pré-requisito: Go instalado.

```bash
go run ./cmd/api
```

Em outro terminal:

```bash
go test ./...
```

## Próximas evoluções

- [ ] PostgreSQL e migrations.
- [ ] Docker Compose para API e banco.
- [ ] Expiração de links.
- [ ] Rate limiting.
- [ ] Observabilidade e CI com GitHub Actions.

## Aprendizados trabalhados

- Design de API HTTP com a biblioteca padrão do Go.
- Separação entre domínio, armazenamento e handlers.
- Concorrência segura com `sync.RWMutex`.
- Testes de unidade e integração de rotas.
