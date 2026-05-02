package order

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/service/order/mocks"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		req model.CreateOrderRq
	}

	type expected struct {
		res      model.CreateOrderRs
		hasErr   bool
		errCheck func(t *testing.T, err error)
	}

	var (
		ctx        = context.Background()
		id         = gofakeit.UUID()
		id2        = gofakeit.UUID()
		id3        = gofakeit.UUID()
		id4        = gofakeit.UUID()
		num        = gofakeit.Int64()
		orderId    = uuid.New()
		orderIdStr = orderId.String()

		req = model.CreateOrderRq{
			HullUUID:   id,
			EngineUUID: id2,
			ShieldUUID: nil,
			WeaponUUID: nil,
		}
		reqOpt = model.CreateOrderRq{
			HullUUID:   id,
			EngineUUID: id2,
			ShieldUUID: &id3,
			WeaponUUID: &id4,
		}

		rqParts    = []string{id, id2}
		rqPartsOpt = []string{id, id2, id3, id4}
		rsParts    = []model.Part{
			{
				Uuid:          id,
				Name:          "hull",
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id2,
				Name:          "engine",
				Price:         num,
				StockQuantity: num,
			},
		}
		rsPartsOut = []model.Part{
			{
				Uuid:          id,
				Name:          "hull",
				Price:         num,
				StockQuantity: 0,
			},
			{
				Uuid:          id2,
				Name:          "engine",
				Price:         num,
				StockQuantity: num,
			},
		}
		rsPartsOpt = []model.Part{
			{
				Uuid:          id,
				Name:          "hull",
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id2,
				Name:          "engine",
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id3,
				Name:          "shield",
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id4,
				Name:          "weapon",
				Price:         num,
				StockQuantity: num,
			},
		}

		errInternal = errors.New("внутренняя ошибка")
	)

	tests := []struct {
		name      string
		args      args
		setupMock func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient)
		expected  expected
	}{
		{
			name: "успешное создание hull+engine - ошибок нет",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient) {
				i.EXPECT().ListParts(ctx, rqParts).Return(rsParts, nil)
				r.EXPECT().Create(ctx, mock.Anything).Return(orderIdStr, nil)
			},
			expected: expected{
				res: model.CreateOrderRs{
					OrderUUID:  orderId,
					TotalPrice: num * int64(len(rsParts)),
				},
				hasErr: false,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "успешное создание shield+weapon - ошибок нет",
			args: args{req: reqOpt},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient) {
				i.EXPECT().ListParts(ctx, rqPartsOpt).Return(rsPartsOpt, nil)
				r.EXPECT().Create(ctx, mock.Anything).Return(orderIdStr, nil)
			},
			expected: expected{
				res: model.CreateOrderRs{
					OrderUUID:  orderId,
					TotalPrice: num * int64(len(rsPartsOpt)),
				},
				hasErr: false,
				errCheck: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
			},
		},
		{
			name: "одна из деталей не найдена - errs.ErrPartNotFound",
			args: args{req: reqOpt},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient) {
				i.EXPECT().ListParts(ctx, rqPartsOpt).Return(rsParts, nil) // rqPartsOpt = 4 записи, rsParts = 2 записи:
				// тк в сервисе сделана проверка на длину слайса в запросе и в ответе ListParts(),
				// если значение в ответе меньше, чем в запросе, то будет выброшена соответствующая ошибка
			},
			expected: expected{
				res:    model.CreateOrderRs{},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrPartNotFound))
				},
			},
		},
		{
			name: "деталь закончилась на складе	- errs.ErrOutOfStock",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient) {
				i.EXPECT().ListParts(ctx, rqParts).Return(rsPartsOut, nil)
			},
			expected: expected{
				res:    model.CreateOrderRs{},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errs.ErrOutOfStock))
				},
			},
		},
		{
			name: "ошибка InventoryService - errInternal",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient) {
				i.EXPECT().ListParts(ctx, rqParts).Return(nil, errInternal)
			},
			expected: expected{
				res:    model.CreateOrderRs{},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
		{
			name: "ошибка при сохранении - errInternal",
			args: args{req: req},
			setupMock: func(r *mocks.OrderRepository, i *mocks.InventoryServiceClient) {
				i.EXPECT().ListParts(ctx, rqParts).Return(rsParts, nil)
				r.EXPECT().Create(ctx, mock.Anything).Return(orderIdStr, errInternal)
			},
			expected: expected{
				res:    model.CreateOrderRs{},
				hasErr: true,
				errCheck: func(t *testing.T, err error) {
					require.Error(t, err)
					assert.True(t, errors.Is(err, errInternal))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := mocks.NewOrderRepository(t)
			i := mocks.NewInventoryServiceClient(t)
			p := mocks.NewPaymentServiceClient(t)
			s := NewService(r, i, p)

			test.setupMock(r, i)

			res, err := s.Create(ctx, test.args.req)

			test.expected.errCheck(t, err)

			if test.expected.hasErr {
				require.Empty(t, res)
			} else {
				require.NotEmpty(t, res)
				assert.Equal(t, test.expected.res.OrderUUID, res.OrderUUID)
				assert.Equal(t, test.expected.res.TotalPrice, res.TotalPrice)
			}
		})
	}
}
