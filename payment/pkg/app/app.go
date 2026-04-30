package app

import (
	"log/slog"

	"google.golang.org/grpc"

	paymentAPIV1 "github.com/MikeKirillov/rocketservices/payment/internal/api/payment/v1"
	orderService "github.com/MikeKirillov/rocketservices/payment/internal/service/order"
	"github.com/MikeKirillov/rocketservices/shared/interceptor"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func RegisterServices(grpcServer *grpc.Server) {
	service := orderService.NewService()
	api := paymentAPIV1.NewAPI(service)
	paymentv1.RegisterPaymentServiceServer(grpcServer, api)
}

func Interceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			sharedErrs.UnaryErrorInterceptor(slog.Default()),
		),
	}
}
