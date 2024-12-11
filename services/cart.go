package services

import (
	"context"
	"errors"

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

	// Check if user has pending orders
	pendingOrders, err := s.orderRepo.FetchPendingOrders(ctx, userID)
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}
	if len(pendingOrders) > 0 {
		return models.PaymentIntentResponse{}, errors.New("user has pending orders")
	}

	// FIXME what if the user has pending orders but a failed payment

	// Reserve cart items
	if err := s.cartRepo.ReserveCartItems(ctx, userID); err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Fetch cart total
	cartTotal, err := s.cartRepo.FetchCartTotal(ctx, userID)
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Create order
	order, err := s.orderRepo.CreateOrder(ctx, userID, cartTotal)
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Send payment request
	paymentIntent, err := s.paymentService.SendPaymentRequest(ctx, models.PaymentIntentRequest{
		Amount:   cartTotal,
		Currency: "usd",
	})
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Save payment intent details
	if err := s.paymentService.SavePayment(ctx, paymentIntent, order.ID); err != nil {
		return models.PaymentIntentResponse{}, err
	}

	// Clear cart
	err = s.cartRepo.ClearCart(ctx, userID)

	return paymentIntent, err
}
