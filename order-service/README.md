# order-service

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
Проверяют SQL, пагинацию, `Cancel` (все ветки: успех, чужой, не `CREATED`, нет заказа), маппинг ошибок БД.

```bash
go test ./internal/adapters/outbound/postgres/... -v
```

**gRPC end-to-end:**  
Поднимают реальный gRPC-сервер (bufconn) + реальный Postgres + реальный Redis. Внешние сервисы (`user-service`, `spot-service`) замоканы через `fx.Decorate`. Проверяют полный цикл: клиент → интерсепторы → хендлер → use case → БД.

```bash
go test ./test/e2e/... -v
```

## Примеры команд

> Токен доступа берётся из `user-service` (метод `Login`). Для админских операций нужен пользователь с ролью `ADMIN`.

### Создание заказа

```bash
ACCESS=<access_token>

# LIMIT-ордер: цена за единицу обязательна
grpcurl -plaintext -H "authorization: Bearer $ACCESS" -d '{
  "market_id": "<MARKET_ID>",
  "side": "ORDER_SIDE_BUY",
  "type": "ORDER_TYPE_LIMIT",
  "price": "60000",
  "quantity": "2"
}' localhost:50053 order.v1.OrderService/CreateOrder

# MARKET-ордер: price не передаётся (игнорируется, если передан)
grpcurl -plaintext -H "authorization: Bearer $ACCESS" -d '{
  "market_id": "<MARKET_ID>",
  "side": "ORDER_SIDE_SELL",
  "type": "ORDER_TYPE_MARKET",
  "quantity": "1"
}' localhost:50053 order.v1.OrderService/CreateOrder
```

Ответ:

```json
{
  "order": {
    "id": "...",
    "userId": "...",
    "marketId": "...",
    "side": "ORDER_SIDE_BUY",
    "type": "ORDER_TYPE_LIMIT",
    "price": "60000",
    "quantity": "2",
    "status": "ORDER_STATUS_CREATED",
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

Если у пользователя нет роли, требуемой рынком (`required_roles` в `spot-service`), заказ создаётся со статусом `ORDER_STATUS_REJECTED` — он сохраняется в истории, но не исполняется.

### Получить заказ

```bash
grpcurl -plaintext -H "authorization: Bearer $ACCESS" -d '{
  "order_id": "<ORDER_ID>"
}' localhost:50053 order.v1.OrderService/GetOrder
```

### Получить только статус заказа

```bash
grpcurl -plaintext -H "authorization: Bearer $ACCESS" -d '{
  "order_id": "<ORDER_ID>"
}' localhost:50053 order.v1.OrderService/GetOrderStatus
```

Ответ:

```json
{
  "orderId": "...",
  "status": "ORDER_STATUS_CREATED",
  "updatedAt": "..."
}
```

### Отмена заказа

```bash
grpcurl -plaintext -H "authorization: Bearer $ACCESS" -d '{
  "order_id": "<ORDER_ID>"
}' localhost:50053 order.v1.OrderService/CancelOrder
```

Отменить можно только свой заказ в статусе `ORDER_STATUS_CREATED`. Повторная отмена → `FailedPrecondition`. `REJECTED`/`FILLED`/`CANCELLED` — терминальные, отмена невозможна.

### Список своих заказов

```bash
grpcurl -plaintext -H "authorization: Bearer $ACCESS" -d '{
  "page_size": 20
}' localhost:50053 order.v1.OrderService/ListOrders
```

Пагинация — cursor-based: в ответе `next_page_token`, передайте его в `page_token` для следующей страницы.

### Админские операции

```bash
ADMIN_ACCESS=<admin_access_token>

# Получить любой заказ
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "order_id": "<ORDER_ID>"
}' localhost:50053 order.v1.OrderService/AdminGetOrder

# Получить только статус любого заказа
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "order_id": "<ORDER_ID>"
}' localhost:50053 order.v1.OrderService/AdminGetOrderStatus

# Отменить чужой заказ
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "order_id": "<ORDER_ID>"
}' localhost:50053 order.v1.OrderService/AdminCancelOrder

# Все заказы (с пагинацией)
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "page_size": 20
}' localhost:50053 order.v1.OrderService/AdminListOrders

# Заказы конкретного пользователя
grpcurl -plaintext -H "authorization: Bearer $ADMIN_ACCESS" -d '{
  "user_id": "<USER_ID>",
  "page_size": 20
}' localhost:50053 order.v1.OrderService/AdminListOrders
```