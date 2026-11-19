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
- Prometheus (в планах)
- **Tempo, Grafana** — мониторинг 

Grafana расположена по адресу localhost:3000. На данный момент дэшборды отсутствуют, только просмотр трейсов через Explore.
Есть трейсинг postgresql и redis запросов.
## Быстрый старт

```bash
cp .env.example .env
docker compose up -d --build
docker compose logs -f <service-name>
```