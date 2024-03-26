package app

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/logger"
	"github.com/pluhe7/gophermart/internal/model"
	"github.com/pluhe7/gophermart/internal/services"
)

type OrderProcessor struct {
	accrualSystem     *services.AccrualSystem
	ordersToProcess   OrdersChanel
	ordersToUpdate    OrdersChanel
	processingCounter int
}

type OrdersChanel struct {
	ordersCh chan *model.Order
}

func (c *OrdersChanel) Add(order *model.Order) {
	c.ordersCh <- order
}

func (c *OrdersChanel) Pop(ctx context.Context) *model.Order {
	select {
	case <-ctx.Done():
	case order := <-c.ordersCh:
		return order
	}

	return nil
}

func NewOrderProcessor(accrualSystemAddress string) *OrderProcessor {
	return &OrderProcessor{
		accrualSystem: services.NewAccrualSystem(accrualSystemAddress),
		ordersToProcess: OrdersChanel{
			make(chan *model.Order, 10),
		},
		ordersToUpdate: OrdersChanel{
			make(chan *model.Order, 10),
		},
	}
}

func (a *App) StartOrderProcessor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	op := a.Server.OrderProcessor

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("order processor done")
			return
		case <-ticker.C:
			op.processingCounter = 0
		default:
			if op.processingCounter >= services.AccrualSystemQueryLimit {
				continue
			}

			order := op.ordersToProcess.Pop(ctx)

			accrualSystemOrder, err := op.accrualSystem.GetOrder(order.Number)
			if err != nil {
				logger.Log.Error("accrual system get order", zap.Error(err))
				continue
			}

			var storageOrderStatus model.OrderStatus

			switch accrualSystemOrder.Status {
			case services.AccrualSystemOrderStatusRegistered, services.AccrualSystemOrderStatusProcessing:
				storageOrderStatus = model.OrderStatusProcessing

			case services.AccrualSystemOrderStatusProcessed:
				storageOrderStatus = model.OrderStatusProcessed

			case services.AccrualSystemOrderStatusInvalid:
				storageOrderStatus = model.OrderStatusInvalid

			default:
				logger.Log.Error("wrong accrual system order status: " + string(accrualSystemOrder.Status))
				continue
			}

			order.Status = storageOrderStatus
			order.Accrual = accrualSystemOrder.Accrual

			op.ordersToUpdate.Add(order)

			op.processingCounter++
		}
	}

}

func (a *App) ProcessOrders(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("orders processed")
				return
			default:
				orders, err := a.Server.Storage.Order().FindUnprocessed()
				if err != nil {
					logger.Log.Error("find unprocessed orders", zap.Error(err))
				}

				for _, order := range orders {
					a.Server.OrderProcessor.ordersToProcess.Add(order)
				}

				time.Sleep(time.Second)
			}
		}
	}()
}

func (a *App) UpdateOrders(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("orders updated")
				return
			default:
				order := a.Server.OrderProcessor.ordersToUpdate.Pop(ctx)

				err := a.Server.Storage.Order().UpdateStatus(order.Number, order.Status)
				if err != nil {
					logger.Log.Error("update status", zap.String("order", order.Number), zap.Error(err))
				}

				if order.Status == model.OrderStatusProcessed {
					if order.Accrual == nil {
						logger.Log.Error("nil accrual", zap.String("order", order.Number))

					} else {
						err = a.Server.Storage.Transaction().Create(&model.Transaction{
							UserID:      order.UserID,
							OrderNumber: order.Number,
							Sum:         *order.Accrual,
							Type:        model.TransactionTypeAccrual,
						})
						if err != nil {
							logger.Log.Error("create transaction", zap.String("order", order.Number), zap.Error(err))
						}

						err = a.Server.Storage.User().UpdateBalance(order.UserID, *order.Accrual)
						if err != nil {
							logger.Log.Error("update user balance", zap.String("order", order.Number), zap.Error(err))
						}
					}
				}
			}
		}
	}()
}
