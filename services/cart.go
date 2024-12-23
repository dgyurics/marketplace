package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CartService interface {
	AddItemToCart(ctx context.Context, item *models.CartItem) error
	GetCart(ctx context.Context) (*models.Cart, error)
	UpdateCartItem(ctx context.Context, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, productID string) error
	ClearCart(ctx context.Context) error
	CheckOut(ctx context.Context) (models.PaymentIntentResponse, error)
}

type cartService struct {
	cartRepo       repositories.CartRepository
	orderRepo      repositories.OrderRepository
	paymentService PaymentService
}

func NewCartService(
	cartRepo repositories.CartRepository,
	orderRepo repositories.OrderRepository,
	paymentService PaymentService) CartService {
	return &cartService{
		cartRepo:       cartRepo,
		orderRepo:      orderRepo,
		paymentService: paymentService,
	}
}

func (s *cartService) AddItemToCart(ctx context.Context, item *models.CartItem) error {
	return s.cartRepo.AddItemToCart(ctx, getUserID(ctx), item)
}

func (s *cartService) GetCart(ctx context.Context) (*models.Cart, error) {
	return s.cartRepo.GetOrCreateCart(ctx, getUserID(ctx))
}

func (s *cartService) UpdateCartItem(ctx context.Context, item *models.CartItem) error {
	return s.cartRepo.UpdateCartItem(ctx, getUserID(ctx), item)
}

func (s *cartService) RemoveItemFromCart(ctx context.Context, productID string) error {
	return s.cartRepo.RemoveItemFromCart(ctx, getUserID(ctx), productID)
}

func (s *cartService) ClearCart(ctx context.Context) error {
	return s.cartRepo.ClearCart(ctx, getUserID(ctx))
}

func (s *cartService) CheckOut(ctx context.Context) (models.PaymentIntentResponse, error) {
	var userID = getUserID(ctx)

	// Create order
	order, err := s.orderRepo.CreateOrder(ctx, userID)
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Send payment request to Stripe
	// On success, this will trigger a webhook event where type = payment_intent.created
	paymentIntent, err := s.paymentService.SendPaymentRequest(ctx, models.PaymentIntentRequest{
		Amount:   order.TotalAmount + order.TaxAmount,
		Currency: "usd",
	})
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Save payment details
	err = s.paymentService.SavePayment(ctx, models.Payment{
		PaymentIntentID: paymentIntent.ID,
		ClientSecret:    paymentIntent.ClientSecret,
		Amount:          paymentIntent.Amount,
		Currency:        paymentIntent.Currency,
		Status:          "pending",
		OrderID:         order.ID,
	})
	return paymentIntent, err
}
