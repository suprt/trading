# Trading

Микросервисная биржевая система на Go.

## Сервисы

- [`user-service`](./user-service/README.md) — аутентификация, пользователи, роли
- [`spot-service`](./spot-service/README.md) — рынки 
- [`order-service`](./order-service/README.md) — заказы 
- `client` — консольный клиент (в планах)

## Общая инфраструктура

- **Postgres** — 3 инстанса (по одному на сервис)
- **Redis** — rate limiting, кэш
- **Prometheus, Jaeger** — мониторинг (в планах)

## Быстрый старт

```bash
cp .env.example .env
docker compose up -d --build
docker compose logs -f <service-name>
```