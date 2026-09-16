package orders

import (
	"context"
	"errors"
	"fmt"

	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/services/order_items"
	"gobrewflow/internal/services/products"

	"github.com/google/uuid"
)

type OrdersService interface {
	CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error)
}

type ordersService struct {
	ordersRepo        OrderRepository
	productsRepo      products.ProductRepository
	orderItemsService order_items.OrderItemsService
}

func NewOrderService(ordersRepo OrderRepository,
	productsRepo products.ProductRepository,
	orderItemsService order_items.OrderItemsService) OrdersService {
	return &ordersService{
		ordersRepo:        ordersRepo,
		productsRepo:      productsRepo,
		orderItemsService: orderItemsService,
	}
}

type CreateOrderInput struct {
	CashierID uuid.UUID
	Items     []order_items.OrderProductItem
}

type CreateOrderOutput struct {
	ID          uuid.UUID
	OrderNumber string
	Status      string

	Items    []order_items.OrderProductItem
	Subtotal int64
	Tax      int64
	Discount int64
	Total    int64

	CashierID uuid.UUID
}

func (s *ordersService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	if input.CashierID == uuid.Nil {
		return nil, ErrInvalidCashierID
	}

	// Load products to get current prices
	productMap := make(map[uuid.UUID]*products.Product, len(input.Items))
	for _, item := range input.Items {
		if item.ProductID == uuid.Nil {
			return nil, products.ErrInvalidProductID
		}

		if item.Quantity <= 0 {
			return nil, inventory.ErrInvalidQuantity
		}

		product, err := s.productsRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		productMap[item.ProductID] = product
	}

	// Calculate item totals and order totals
	var subtotal, tax, discount, total int64
	for i := range input.Items {

		product := productMap[input.Items[i].ProductID]
		input.Items[i].UnitPrice = product.Price
		input.Items[i].Total = product.Price * int64(input.Items[i].Quantity)
		subtotal += input.Items[i].Total
	}

	total = subtotal + tax - discount

	// Begin transaction
	tx, err := s.ordersRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create order
	order := &Orders{
		Status:    OrderStatusPending,
		Subtotal:  subtotal,
		Tax:       tax,
		Discount:  discount,
		Total:     total,
		CashierID: input.CashierID,
	}

	if err := s.ordersRepo.InsertOrder(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	orderItemsInput := &order_items.CreateOrderItemsInput{
		OrderID: order.ID,
		Items:   input.Items,
	}

	if err := s.orderItemsService.CreateOrderItems(ctx, tx, orderItemsInput); err != nil {
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CreateOrderOutput{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		Status:      string(order.Status),
		Items:       input.Items,
		Subtotal:    subtotal,
		Tax:         tax,
		Discount:    discount,
		Total:       total,
		CashierID:   order.CashierID,
	}, nil
}
