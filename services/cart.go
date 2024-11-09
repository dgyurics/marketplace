package services

import (
	"context"
	"sync"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CartService interface {
	AddItemToCart(ctx context.Context, item *models.CartItem) error
	GetCart(ctx context.Context) (*models.Cart, error)
	UpdateCartItem(ctx context.Context, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, productID string) error
	ClearCart(ctx context.Context) error
	CheckOut(ctx context.Context, tokenID string) (models.PaymentIntentResponse, error)
}

type cartService struct {
	repo           repositories.CartRepository
	paymentService PaymentService
}

func NewCartService(
	repo repositories.CartRepository,
	paymentService PaymentService) CartService {
	return &cartService{
		repo:           repo,
		paymentService: paymentService,
	}
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

func (s *cartService) CheckOut(ctx context.Context, tokenID string) (models.PaymentIntentResponse, error) {
	var userID = getUserID(ctx)
	var wg sync.WaitGroup
	var cartTotal models.Currency
	errChan := make(chan error, 2)

	// Reserve cart items
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.repo.ReserveCartItems(ctx, userID); err != nil {
			errChan <- err
		}
	}()

	// Fetch cart total
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		// TODO schedule a job to release the reserved inventory after a certain time
		// TODO prevent the user from reserving multiple times
		// TODO include taxes and shipping
		cartTotal, err = s.repo.FetchCartTotal(ctx, userID)
		if err != nil {
			errChan <- err
		}
	}()

	// Wait for goroutines to finish
	wg.Wait()
	close(errChan)

	// return the first error encountered, if any
	for err := range errChan {
		return models.PaymentIntentResponse{}, err
	}

	// Create a Payment Intent with a third-party payment processor
	tmp := models.PaymentIntentRequest{
		Provider: models.Stripe,
		Amount:   cartTotal,
		Currency: "usd",
		TokenID:  tokenID,
	}
	// TODO save res details to payment_transactions table
	return s.paymentService.SendPaymentRequest(tmp)
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
