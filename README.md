<!-- TODO: Замени YOUR_USERNAME на свой GitHub username и YOUR_GIST_ID на ID своего gist для coverage badge -->
![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/YOUR_USERNAME/YOUR_GIST_ID/raw/coverage.json)

# server calls examples

## payment

### server start
```
$ go run cmd/grpc_server/main.go
```
если файл `.env` (в данном случае `payment.env`) отсутствует, то по умолчанию будет определяться значение пути 
к файлу с конфигом по умолчанию - в коде. если установлен, то будет браться из него, если не переданы дополнительные
атрибуты при старте приложения, как в примерах ниже:

* старт приложения с передачей env пути к файлу с конфигом и gRPC-порта
```
$ CONFIG_PATH=config.production.yaml GRPC_PORT=50051 go run cmd/grpc_server/main.go
```
* старт приложения с передачей флага пути к файлу с конфигом
```
$ go run cmd/grpc_server/main.go -config=config.production.yaml
```
при этом приоритет подгрузки параметров для текущего приложения, исходя из реализации, следующий:
`env` в команде старта > `flag` в команде старта > `.env`-file > `.yaml`-file > если ничего не определено, то
дефолтный файл, если указан, или дефолтные значения `env`

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
$ go run cmd/grpc_server/main.go
```

### получить все детали
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{}' localhost:50051 inventory.v1.InventoryService/ListParts
```

### получить только корпуса (фильтр по типу)
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"part_type": "PART_TYPE_HULL"}' localhost:50051 inventory.v1.InventoryService/ListParts
```

### получить конкретные детали по UUID (batch)
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"uuids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440003", "550e8400-e29b-41d4-a716-446655440005"]}' localhost:50051 inventory.v1.InventoryService/ListParts
```

### получить одну деталь по UUID
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"uuid": "550e8400-e29b-41d4-a716-446655440001"}' localhost:50051 inventory.v1.InventoryService/GetPart
```

### провалидировать детали по UUID
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"hull_uuid": "550e8400-e29b-41d4-a716-446655440002", "engine_uuid": "550e8400-e29b-41d4-a716-446655440004"}' localhost:50051 inventory.v1.InventoryService/ValidateCompatibility
```

### зарезервировать деталь по UUID
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"uuids": ["550e8400-e29b-41d4-a716-446655440001"]}' localhost:50051 inventory.v1.InventoryService/ReserveParts
```

### отменить резерв детали по UUID
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"uuids": ["550e8400-e29b-41d4-a716-446655440001"]}' localhost:50051 inventory.v1.InventoryService/ReleaseParts
```

### коммит деталей по UUID
```
$ grpcurl -plaintext -H "session-uuid: 37bde4d1-23de-49ff-85c0-75587407d9f5" -d '{"uuids": ["550e8400-e29b-41d4-a716-446655440001"]}' localhost:50051 inventory.v1.InventoryService/CommitParts
```

## order

### server start
```
$ go run cmd/main.go
```

### создание заказа
```
$ curl -v -X POST http://localhost:8080/api/v1/orders -H "Content-Type: application/json" -H "Authorization: Bearer 86926bfc-87a6-432e-9b58-fefea8acd47a" -d '{"hull_uuid": "550e8400-e29b-41d4-a716-446655440001", "engine_uuid": "550e8400-e29b-41d4-a716-446655440003", "shield_uuid": "550e8400-e29b-41d4-a716-446655440005", "weapon_uuid": null}'
```

### получить заказ
```
$ curl -v http://localhost:8080/api/v1/orders/8c8e205f-f2f9-46be-a702-9cf0ca3298b9 -H "Authorization: Bearer 86926bfc-87a6-432e-9b58-fefea8acd47a"
```

### оплата заказа
```
$ curl -v -X POST http://localhost:8080/api/v1/orders/8c8e205f-f2f9-46be-a702-9cf0ca3298b9/pay -H "Content-Type: application/json" -H "Authorization: Bearer 86926bfc-87a6-432e-9b58-fefea8acd47a" -d '{"payment_method": "CARD"}'
```

### отмена заказа
```
$ curl -v -X POST http://localhost:8080/api/v1/orders/8c8e205f-f2f9-46be-a702-9cf0ca3298b9/cancel -H "Content-Type: application/json" -H "Authorization: Bearer 86926bfc-87a6-432e-9b58-fefea8acd47a"
```

## assembly

### server start
```
$ go run cmd/main.go
```

## iam

### server start
```
$ go run cmd/grpc_server/main.go
```

### зарегистрировать пользователя
```
$ grpcurl -plaintext -d '{"info": {"info": {"login": "test"}, "password": "password123"}}' localhost:50053 user.v1.UserService/Register
```

### получить пользователя
```
$ grpcurl -plaintext -d '{"userUuid": "62ba15d3-e7d3-478f-852c-77373540454b"}' localhost:50053 user.v1.UserService/GetUser
```

### логин
```
$ grpcurl -plaintext -d '{"login": "test", "password": "password123"}' localhost:50053 auth.v1.AuthService/Login
```

### логаут
```
$ grpcurl -plaintext -d '{"session_uuid": "50be5582-8512-43c6-b958-44080c86da6d"}' localhost:50053 auth.v1.AuthService/Logout
```

### о сессии и пользователе
```
$ grpcurl -plaintext -d '{"session_uuid": "50be5582-8512-43c6-b958-44080c86da6d"}' localhost:50053 auth.v1.AuthService/Whoami
```

# goose файл миграции

```
$ bin/goose -dir migrations/inventory create create_parts_table sql
$ bin/goose -dir migrations/inventory create seed_parts sql
$ bin/goose -dir migrations/inventory create add_properties_reserved sql
$ bin/goose -dir migrations/inventory create update_seed_properties sql
$ bin/goose -dir migrations/inventory create increase_stock_quantity sql

$ bin/goose -dir migrations/order create create_orders_table sql
$ bin/goose -dir migrations/order create create_order_items_table sql
$ bin/goose -dir migrations/order create add_user_uuid_column_to_orders sql
$ bin/goose -dir migrations/order create backfill_user_uuid_in_orders sql
$ bin/goose -dir migrations/order create set_user_uuid_not_null_in_orders sql

$ bin/goose -dir migrations/iam create create_users_table sql
```

# Kafka UI

```
http://localhost:8090/
```

# Kibana - логирование

сначала поднять все окружение `task up-all` + миграции `task migrate-all-up`, затем выполнить `task kibana:init`

```
http://localhost:5601/
```

# Grafana - метрики

login/pass: admin/admin

```
http://localhost:3000/
```

# Jaeger - трейсы

```
http://localhost:16686/
```
