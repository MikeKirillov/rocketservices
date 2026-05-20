<!-- TODO: Замени YOUR_USERNAME на свой GitHub username и YOUR_GIST_ID на ID своего gist для coverage badge -->
![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/YOUR_USERNAME/YOUR_GIST_ID/raw/coverage.json)

# server calls examples

## payment

### server start
```
$ go run payment/cmd/grpc_server/main.go
```

### провести оплату
```
$ grpcurl -plaintext -d '{
  "order_uuid": "550e8400-e29b-41d4-a716-000000000001",
  "payment_method": "PAYMENT_METHOD_CARD"
}' localhost:50052 payment.v1.PaymentService/PayOrder
```

## inventory

### server start
```
$ go run inventory/cmd/grpc_server/main.go
```

### получить все детали
```
$ grpcurl -plaintext -d '{}' localhost:50051 inventory.v1.InventoryService/ListParts
```

### получить только корпуса (фильтр по типу)
```
$ grpcurl -plaintext -d '{"part_type": "PART_TYPE_HULL"}' localhost:50051 inventory.v1.InventoryService/ListParts
```

### получить конкретные детали по UUID (batch)
```
$ grpcurl -plaintext -d '{"uuids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440003", "550e8400-e29b-41d4-a716-446655440005"]}' localhost:50051 inventory.v1.InventoryService/ListParts
```

### получить одну деталь по UUID
```
$ grpcurl -plaintext -d '{"uuid": "550e8400-e29b-41d4-a716-446655440001"}' localhost:50051 inventory.v1.InventoryService/GetPart
```

## order

### server start
```
$ go run order/cmd/main.go
```

### создание заказа
```
$ curl -v -X POST http://localhost:8080/api/v1/orders -H "Content-Type: application/json" -d '{"hull_uuid": "550e8400-e29b-41d4-a716-446655440001", "engine_uuid": "550e8400-e29b-41d4-a716-446655440003", "shield_uuid": "550e8400-e29b-41d4-a716-446655440005", "weapon_uuid": null}'
```

### получить заказ
```
$ curl -v http://localhost:8080/api/v1/orders/{order_uuid}
```

### оплата заказа
```
$ curl -v -X POST http://localhost:8080/api/v1/orders/{order_uuid}/pay -H "Content-Type: application/json" -d '{"payment_method": "CARD"}'
```

### отмена заказа
```
$ curl -v -X POST http://localhost:8080/api/v1/orders/{order_uuid}/cancel -H "Content-Type: application/json"
```

# goose файл миграции

```
$ bin/goose -dir migrations/inventory create create_parts_table sql
$ bin/goose -dir migrations/inventory create seed_parts sql
$ bin/goose -dir migrations/order create create_orders_table sql
$ bin/goose -dir migrations/order create create_order_items_table sql
```