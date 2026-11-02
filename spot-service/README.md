# spot-service

## Тесты

### Юнит-тесты

```bash
# use cases
go test ./internal/application/... -v

# gRPC-хендлеры (mockgen)
go test ./internal/adapters/inbound/grpc/... -v
```

### Интеграционные (testcontainers)

**Репозитории (Postgres):**  
Проверяют SQL, транзакции, маппинг ошибок БД.

```bash
go test ./internal/adapters/outbound/postgres/... -v
```

**gRPC end-to-end:**  
Поднимают реальный gRPC-сервер (bufconn) + реальный Postgres + реальный Redis. Проверяют полный цикл: клиент → интерсепторы → хендлер → use case → БД.

```bash
go test ./test/e2e/... -v
```

## Примеры команд

> Токен доступа берётся из `user-service` (метод `Login`). Для админских операций нужен пользователь с ролью `ADMIN`.

### Просмотр доступных рынков (без авторизации)

```bash
grpcurl -plaintext -d '{
  "user_roles": ["ROLE_USER"]
}' localhost:50052 spot.v1.SpotInstrumentService/ViewMarkets
```

Ответ:

```json
{
  "markets": [
    {
      "marketId": "...",
      "name": "BTC/USDT",
      "enabled": true,
      "createdAt": "...",
      "updatedAt": "..."
    }
  ]
}
```

### Получить рынок по ID (без авторизации)

```bash
grpcurl -plaintext -d '{
  "market_id": "<MARKET_ID>"
}' localhost:50052 spot.v1.SpotInstrumentService/GetMarket
```

### Создание рынка (только ADMIN)

```bash
ADMIN_ACCESS=<admin_access_token>

grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "name": "ETH/USDT",
  "required_roles": ["ROLE_USER"],
  "enabled": true
}' localhost:50052 spot.v1.SpotInstrumentService/CreateMarket
```

Ответ:

```json
{
  "market": {
    "marketId": "...",
    "name": "ETH/USDT",
    "requiredRoles": ["ROLE_USER"],
    "enabled": true
  }
}
```

### Обновление деталей рынка (только ADMIN)

```bash
# Меняем только name
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "market_id": "<MARKET_ID>",
  "name": "ETH/USDC"
}' localhost:50052 spot.v1.SpotInstrumentService/UpdateMarketDetails

# Меняем только enabled
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "market_id": "<MARKET_ID>",
  "enabled": false
}' localhost:50052 spot.v1.SpotInstrumentService/UpdateMarketDetails
```

### Обновление ролей рынка (только ADMIN)

```bash
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "market_id": "<MARKET_ID>",
  "required_roles": ["ROLE_TRADER"]
}' localhost:50052 spot.v1.SpotInstrumentService/UpdateMarketRoles
```

### Удаление рынка (soft-delete, только ADMIN)

```bash
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "market_id": "<MARKET_ID>"
}' localhost:50052 spot.v1.SpotInstrumentService/DeleteMarket
```

### Список всех рынков (только ADMIN)

```bash
# Все, включая удалённые и выключенные
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "include_deleted": true,
  "include_disabled": true
}' localhost:50052 spot.v1.SpotInstrumentService/ListAllMarkets

# Только активные
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{}' \
  localhost:50052 spot.v1.SpotInstrumentService/ListAllMarkets
```