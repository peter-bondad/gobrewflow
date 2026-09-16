package order_items

import (
	"context"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/services/products"
	"time"

	"github.com/google/uuid"
)

type OrderItemsService interface {
	CreateOrderItems(ctx context.Context, input *CreateOrderItemsInput) error
}

type orderItemsService struct {
	orderItemRepository OrderItemRepository
}

func NewOrderItemsService(orderItemRepository OrderItemRepository) OrderItemsService {
	return &orderItemsService{
		orderItemRepository: orderItemRepository,
	}
}

type CreateOrderItemsInput struct {
	OrderID uuid.UUID
	Items   []OrderProductItem
}

type OrderProductItem struct {
	ProductID uuid.UUID
	Quantity  int
	UnitPrice int64
	Total     int64
}

func (s *orderItemsService) CreateOrderItems(ctx context.Context, input *CreateOrderItemsInput) error {
	for _, item := range input.Items {
		if err := validateCreateOrderInput(item.ProductID, item.Quantity, item.UnitPrice); err != nil {
			return err
		}

		orderItem := &OrderItem{
			ID:        uuid.New(),
			OrderID:   input.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Total:     item.Total,
			CreatedAt: time.Now(),
		}

		if err := s.orderItemRepository.InsertOrderItem(ctx, orderItem); err != nil {
			return err
		}
	}

	return nil
}

func validateCreateOrderInput(
	productID uuid.UUID,
	quantity int,
	unitPrice int64,
) error {

	if productID == uuid.Nil {
		return products.ErrInvalidProductID
	}

	if quantity <= 0 {
		return inventory.ErrInvalidQuantity
	}

	if unitPrice < 0 {
		return products.ErrInvalidUnitPrice
	}

	return nil
}
