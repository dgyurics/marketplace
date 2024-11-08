package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CartService interface {
	CreateCart(ctx context.Context, cart *models.Cart) error
	AddItemToCart(ctx context.Context, item *models.CartItem) error
	GetCart(ctx context.Context) (*models.Cart, error)
	UpdateCartItem(ctx context.Context, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, productID string) error
	ClearCart(ctx context.Context) error
	CheckOut(ctx context.Context) error
}

type cartService struct {
	repo repositories.CartRepository
}

func NewCartService(repo repositories.CartRepository) CartService {
	return &cartService{repo: repo}
}

func (s *cartService) CreateCart(ctx context.Context, cart *models.Cart) error {
	return s.repo.CreateCart(ctx, getUserID(ctx))
}

func (s *cartService) AddItemToCart(ctx context.Context, item *models.CartItem) error {
	return s.repo.AddItemToCart(ctx, getUserID(ctx), item)
}

func (s *cartService) GetCart(ctx context.Context) (*models.Cart, error) {
	return s.repo.GetOrCreateCart(ctx, getUserID(ctx))
}

func (s *cartService) UpdateCartItem(ctx context.Context, item *models.CartItem) error {
	return s.repo.UpdateCartItem(ctx, getUserID(ctx), item)
}

func (s *cartService) RemoveItemFromCart(ctx context.Context, productID string) error {
	return s.repo.RemoveItemFromCart(ctx, getUserID(ctx), productID)
}

func (s *cartService) ClearCart(ctx context.Context) error {
	return s.repo.ClearCart(ctx, getUserID(ctx))
}

func (s *cartService) CheckOut(ctx context.Context) error {
	// Temporarily hold the inventory for items in the cart to prevent others
	// from purchasing the same stock.
	err := s.repo.ReserveCartItems(ctx, getUserID(ctx))

	// TODO schedule a job to release the reserved inventory after a certain time
	// TODO prevent the user from reserving multiple times

	// 2. Calculate Total
	//    - Calculate the final total for the cart, including taxes, shipping, and any discounts.

	// 3. Initiate Payment Intent
	//    - Create a Payment Intent with a third-party payment processor (e.g., Stripe) for the calculated total.
	//    - Return the Payment Intent’s client secret to the front end to allow the customer to complete payment.

	// After payment intent is created, payment confirmation will call CompleteOrder
	return err
}

func (s *cartService) CompleteOrder(ctx context.Context, paymentStatus string) error {
	// 5. Confirm Payment Status
	//    - If the payment succeeds:
	//       - Record the transaction details in a `payment_transactions` table (provider, transaction ID, amount, etc.).
	//       - Proceed to order creation.
	//    - If payment fails, release the reserved inventory and notify the user.

	// 6. Create Order
	//    - Create an order record in the `orders` table, linking it with the user, payment, and shipping details.
	//    - Populate `order_items` with each cart item and its relevant pricing data.

	// 7. Deduct Inventory
	//    - Permanently reduce inventory for each item based on the final order quantities.

	// 8. Clear Cart
	//    - Clear or reset the cart for future purchases, ensuring it’s ready for the next session.

	// 9. Return Success Response
	//    - Notify the user that the order was successfully placed.

	return nil
}
