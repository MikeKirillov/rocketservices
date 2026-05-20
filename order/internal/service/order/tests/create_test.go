package tests

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
	service "github.com/MikeKirillov/rocketservices/order/internal/service/order"
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
		ctx     = context.Background()
		id      = uuid.New()
		id2     = uuid.New()
		id3     = uuid.New()
		id4     = uuid.New()
		num     = gofakeit.Int64()
		orderId = uuid.New()
		items   = []model.OrderItem{
			{Uuid: id, Type: model.PartTypeHull},
			{Uuid: id2, Type: model.PartTypeEngine},
		}
		itemsOpt = []model.OrderItem{
			{Uuid: id, Type: model.PartTypeHull},
			{Uuid: id2, Type: model.PartTypeEngine},
			{Uuid: id3, Type: model.PartTypeShield},
			{Uuid: id4, Type: model.PartTypeWeapon},
		}

		req = model.CreateOrderRq{
			Items: items,
		}
		reqOpt = model.CreateOrderRq{
			Items: itemsOpt,
		}

		rqItems    = []uuid.UUID{id, id2}
		rqItemsOpt = []uuid.UUID{id, id2, id3, id4}
		rsItems    = []model.OrderItem{
			{
				Uuid:          id,
				Name:          new("hull"),
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id2,
				Name:          new("engine"),
				Price:         num,
				StockQuantity: num,
			},
		}
		rsPartsOut = []model.OrderItem{
			{
				Uuid:          id,
				Type:          model.PartTypeHull,
				Name:          new("hull"),
				Price:         num,
				StockQuantity: 0,
			},
			{
				Uuid:          id2,
				Type:          model.PartTypeEngine,
				Name:          new("engine"),
				Price:         num,
				StockQuantity: num,
			},
		}
		rsPartsOpt = []model.OrderItem{
			{
				Uuid:          id,
				Type:          model.PartTypeHull,
				Name:          new("hull"),
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id2,
				Type:          model.PartTypeEngine,
				Name:          new("engine"),
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id3,
				Type:          model.PartTypeShield,
				Name:          new("shield"),
				Price:         num,
				StockQuantity: num,
			},
			{
				Uuid:          id4,
				Type:          model.PartTypeWeapon,
				Name:          new("weapon"),
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
				i.EXPECT().ListParts(ctx, rqItems).Return(rsItems, nil)
				i.EXPECT().ValidateCompatibility(ctx, rsItems).Return(nil)
				i.EXPECT().ReserveParts(ctx, rqItems).Return(nil)
				r.EXPECT().Create(ctx, mock.MatchedBy(func(o model.Order) bool {
					return o.OrderUUID != uuid.Nil &&
						o.Items != nil &&
						len(o.Items) > 0 &&
						o.TotalPrice != 0 &&
						o.Status == model.OrderStatusPendingPayment &&
						!o.CreatedAt.IsZero()
				})).Return(nil)
			},
			expected: expected{
				res: model.CreateOrderRs{
					OrderUUID:  orderId,
					TotalPrice: num * int64(len(rsItems)),
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
				i.EXPECT().ListParts(ctx, rqItemsOpt).Return(rsPartsOpt, nil)
				i.EXPECT().ValidateCompatibility(ctx, rsPartsOpt).Return(nil)
				i.EXPECT().ReserveParts(ctx, rqItemsOpt).Return(nil)
				r.EXPECT().Create(ctx, mock.MatchedBy(func(o model.Order) bool {
					return o.OrderUUID != uuid.Nil &&
						o.Items != nil &&
						len(o.Items) > 0 &&
						o.TotalPrice != 0 &&
						o.Status == model.OrderStatusPendingPayment &&
						!o.CreatedAt.IsZero()
				})).Return(nil)
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
				i.EXPECT().ListParts(ctx, rqItemsOpt).Return(rsItems, nil) // rqPartsOpt = 4 записи, rsParts = 2 записи:
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
				i.EXPECT().ListParts(ctx, rqItems).Return(rsPartsOut, nil)
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
				i.EXPECT().ListParts(ctx, rqItems).Return(nil, errInternal)
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
				i.EXPECT().ListParts(ctx, rqItems).Return(rsItems, nil)
				i.EXPECT().ValidateCompatibility(ctx, rsItems).Return(nil)
				i.EXPECT().ReserveParts(ctx, rqItems).Return(nil)
				r.EXPECT().Create(ctx, mock.MatchedBy(func(o model.Order) bool {
					return o.OrderUUID != uuid.Nil &&
						o.Items != nil &&
						len(o.Items) > 0 &&
						o.TotalPrice != 0 &&
						o.Status == model.OrderStatusPendingPayment &&
						!o.CreatedAt.IsZero()
				})).Return(errInternal)
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
			m := mocks.NewTxManager(t)
			s := service.NewService(r, i, p, m)

			test.setupMock(r, i)

			res, err := s.Create(ctx, test.args.req)

			test.expected.errCheck(t, err)

			if test.expected.hasErr {
				require.Empty(t, res)
			} else {
				require.NotEmpty(t, res)
				assert.NotNil(t, res.OrderUUID)
				assert.Equal(t, test.expected.res.TotalPrice, res.TotalPrice)
			}
		})
	}
}
