package v1

import (
	"time"

	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

// Таймаут для gRPC вызовов.
const grpcCallTimeout = 5 * time.Second

type client struct {
	grpcClient inventoryv1.InventoryServiceClient
}

// New создает обертку над gRPC клиентом InventoryService.
func New(grpcClient inventoryv1.InventoryServiceClient) *client {
	return &client{grpcClient: grpcClient}
}
