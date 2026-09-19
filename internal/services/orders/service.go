package orders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gobrewflow/internal/database"
	"gobrewflow/internal/services/inventory"
	"gobrewflow/internal/services/inventory_movements"
	"gobrewflow/internal/services/order_items"
	"gobrewflow/internal/services/products"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OrdersService interface {
	CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error)
}

type ordersService struct {
	ordersRepo             OrderRepository
	productsRepo           products.ProductRepository
	inventoryRepo          inventory.InventoryRepository
	inventoryMovementsRepo inventory_movements.InventoryMovementsRepository
	orderItemsService      order_items.OrderItemsService
	txManager              database.TxManager
}

func NewOrderService(
	ordersRepo OrderRepository,
	productsRepo products.ProductRepository,
	inventoryRepo inventory.InventoryRepository,
	inventoryMovementsRepo inventory_movements.InventoryMovementsRepository,
	orderItemsService order_items.OrderItemsService,
) OrdersService {
	return &ordersService{
		ordersRepo:             ordersRepo,
		productsRepo:           productsRepo,
		inventoryRepo:          inventoryRepo,
		inventoryMovementsRepo: inventoryMovementsRepo,
		orderItemsService:      orderItemsService,
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

type InsufficientStockItem struct {
	ProductID    uuid.UUID
	RequestedQty int
	AvailableQty int
}

type InsufficientStockError struct {
	Items []InsufficientStockItem
}

func (e *InsufficientStockError) Error() string {
	return "one or more products have insufficient stock"
}

func (s *ordersService) CreateOrder(
	ctx context.Context,
	input *CreateOrderInput,
) (*CreateOrderOutput, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	if input.CashierID == uuid.Nil {
		return nil, ErrInvalidCashierID
	}

	// Aggregate requested quantities by product.
	requestedQty := make(map[uuid.UUID]int, len(input.Items))

	for _, item := range input.Items {
		if item.ProductID == uuid.Nil {
			return nil, products.ErrInvalidProductID
		}

		if item.Quantity <= 0 {
			return nil, inventory.ErrInvalidQuantity
		}

		requestedQty[item.ProductID] += item.Quantity
	}

	// Load products and validate inventory.
	productMap := make(map[uuid.UUID]*products.ProductListItem, len(requestedQty))
	var insufficientStock []InsufficientStockItem

	for productID, quantity := range requestedQty {
		product, err := s.productsRepo.FindByID(ctx, productID)
		if err != nil {
			return nil, err
		}

		stock, err := s.inventoryRepo.FindByProductID(ctx, productID)
		if err != nil {
			return nil, err
		}

		if stock.Quantity < quantity {
			insufficientStock = append(insufficientStock, InsufficientStockItem{
				ProductID:    productID,
				RequestedQty: quantity,
				AvailableQty: stock.Quantity,
			})
			continue
		}

		productMap[productID] = product
	}

	if len(insufficientStock) > 0 {
		return nil, &InsufficientStockError{
			Items: insufficientStock,
		}
	}

	// Calculate item totals and order totals.
	var subtotal, tax, discount, total int64

	for i := range input.Items {
		product := productMap[input.Items[i].ProductID]

		input.Items[i].UnitPrice = product.Price
		input.Items[i].Total =
			product.Price * int64(input.Items[i].Quantity)

		subtotal += input.Items[i].Total
	}

	total = subtotal + tax - discount

	var order *Orders

	err := s.txManager.WithTx(ctx, func(tx bun.IDB) error {
		// Deduct stock and create inventory movements.
		for productID, quantity := range requestedQty {
			_, err := s.inventoryRepo.ChangeStock(
				ctx,
				tx,
				inventory.StockParams{
					ProductID: productID,
					Quantity:  quantity,
					Change:    inventory.StockDecrease,
				},
			)
			if err != nil {
				return err
			}

			movement := &inventory_movements.InventoryMovement{
				ID:        uuid.New(),
				ProductID: productID,
				Type:      inventory_movements.InventoryMovementTypeSold,
				Quantity:  quantity,
				CreatedAt: time.Now(),
			}

			if err := s.inventoryMovementsRepo.CreateMovement(
				ctx,
				tx,
				movement,
			); err != nil {
				return fmt.Errorf("failed to create inventory movement: %w", err)
			}
		}

		// Create order.
		order = &Orders{
			Status:    OrderStatusPending,
			Subtotal:  subtotal,
			Tax:       tax,
			Discount:  discount,
			Total:     total,
			CashierID: input.CashierID,
		}

		if err := s.ordersRepo.InsertOrder(ctx, tx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Create order items using the same transaction.
		orderItemsInput := &order_items.CreateOrderItemsInput{
			OrderID: order.ID,
			Items:   input.Items,
		}

		if err := s.orderItemsService.CreateOrderItems(
			ctx,
			tx,
			orderItemsInput,
		); err != nil {
			return fmt.Errorf("failed to create order items: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
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
