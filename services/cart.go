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
	ConfirmPayment(ctx context.Context, paymentIntentID string) error
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

	// TODO verify account has address information

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
	paymentIntent := models.PaymentIntentRequest{
		Provider: models.Stripe,
		Amount:   cartTotal,
		Currency: "usd",
		TokenID:  tokenID,
	}

	// TODO implement retry logic for network failures, timeouts, etc.
	response, err := s.paymentService.SendPaymentRequest(ctx, paymentIntent)
	if err != nil || response.Status != "success" {
		return response, err
	}

	// Start the order
	err = s.repo.StartOrder(ctx, userID, response.PaymentIntentID, response.Amount)
	return response, err
}

func (s *cartService) ConfirmPayment(ctx context.Context, paymentIntentID string) error {
	// fetch payment intent details
	payIntent, err := s.paymentService.RetrievePaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}

	// fetch expected cart total
	cartTotal, err := s.repo.FetchCartTotal(ctx, getUserID(ctx))
	if err != nil {
		return err
	}

	// compare payment amount with cart total
	if payIntent.AmountReceived != cartTotal {
		// TODO: refund the payment
		// TODO: log the discrepancy and generate an alert
		// return errors.New("payment amount does not match cart total")
	}

	// process the order
	return s.repo.CompleteOrder(ctx, getUserID(ctx), paymentIntentID)
}
