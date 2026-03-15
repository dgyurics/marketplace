package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type PurchaseIntentService interface {
	CreatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error
	UpdatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error
	GetPurchaseIntentByID(ctx context.Context, id string) (types.PurchaseIntent, error)
	GetPurchaseIntentsByProductID(ctx context.Context, id string) ([]types.PurchaseIntent, error)
	GetPurchaseIntents(ctx context.Context) ([]types.PurchaseIntent, error)
}

func NewPurchaseIntentService(
	repoProduct repositories.ProductRepository,
	repoPurchaseIntent repositories.PurchaseIntentRepository,
	serviceUser UserService,
	serviceProduct ProductService,
	templateService TemplateService,
	emailService EmailService,
	baseURL string,
) PurchaseIntentService {
	return &purchaseIntentService{
		repoPurchaseIntent: repoPurchaseIntent,
		repoProduct:        repoProduct,
		serviceUser:        serviceUser,
		serviceProduct:     serviceProduct,
		templateService:    templateService,
		emailService:       emailService,
		baseURL:            baseURL,
	}
}

type purchaseIntentService struct {
	repoPurchaseIntent repositories.PurchaseIntentRepository
	repoProduct        repositories.ProductRepository
	serviceUser        UserService
	serviceProduct     ProductService
	templateService    TemplateService
	emailService       EmailService
	baseURL            string
}

func (ps *purchaseIntentService) CreatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) (err error) {
	purchaseIntent.UserID = getUserID(ctx)
	purchaseIntent.ID, err = utilities.GenerateIDString()
	if err != nil {
		return err
	}

	// Determine status based on offer price and product price
	product, err := ps.repoProduct.GetProductByID(ctx, purchaseIntent.Product.ID)
	if err != nil {
		return err
	}
	if product.Price <= purchaseIntent.OfferPrice {
		purchaseIntent.Status = types.PurchaseIntentAccepted
	} else {
		purchaseIntent.Status = types.PurchaseIntentPending
	}

	if err := ps.repoPurchaseIntent.CreatePurchaseIntent(ctx, purchaseIntent); err != nil {
		return err
	}

	go ps.PaymentIntentNotificationEmail(*purchaseIntent)

	// email customer if auto-accepted
	if purchaseIntent.Status == types.PurchaseIntentAccepted {
		purchaseIntent.Product, err = ps.serviceProduct.GetProductByID(ctx, purchaseIntent.Product.ID)
		if err != nil {
			return err
		}
		go ps.PaymentIntentUpdateEmail(*purchaseIntent)
	}

	return nil
}

func (ps *purchaseIntentService) UpdatePurchaseIntent(ctx context.Context, purchaseIntent *types.PurchaseIntent) error {
	if err := ps.repoPurchaseIntent.UpdatePurchaseIntent(ctx, purchaseIntent); err != nil {
		return err
	}

	// email customer if purchaseintent status changes to accepted or rejected
	if purchaseIntent.Status == types.PurchaseIntentAccepted || purchaseIntent.Status == types.PurchaseIntentRejected {
		product, err := ps.serviceProduct.GetProductByID(ctx, purchaseIntent.Product.ID)
		if err != nil {
			return err
		}
		purchaseIntent.Product = product
		go ps.PaymentIntentUpdateEmail(*purchaseIntent)
	}
	return nil
}

func (ps *purchaseIntentService) GetPurchaseIntentsByProductID(ctx context.Context, id string) ([]types.PurchaseIntent, error) {
	return ps.repoPurchaseIntent.GetPurchaseIntentsByProductIDAndUser(ctx, id, getUserID(ctx))
}

func (ps *purchaseIntentService) GetPurchaseIntentByID(ctx context.Context, id string) (types.PurchaseIntent, error) {
	return ps.repoPurchaseIntent.GetPurchaseIntentByID(ctx, id)
}

func (ps *purchaseIntentService) GetPurchaseIntents(ctx context.Context) ([]types.PurchaseIntent, error) {
	return ps.repoPurchaseIntent.GetPurchaseIntents(ctx)
}

// Send payment intent update to user
func (ps *purchaseIntentService) PaymentIntentUpdateEmail(purchaseIntent types.PurchaseIntent) {
	usr, err := ps.serviceUser.GetUserByID(context.Background(), purchaseIntent.UserID)
	if err != nil {
		slog.Error("Error fetching user", "ID", purchaseIntent.UserID, "error", err)
		return
	}

	data := map[string]string{
		"ProductName": purchaseIntent.Product.Name,
		"Status":      string(purchaseIntent.Status),
	}
	body, err := ps.templateService.RenderToString(PurchaseIntentUpdate, data)
	if err != nil {
		slog.Error("Error loading email template: ", "error", err)
		return
	}
	email := &types.Email{
		To:      []string{usr.Email},
		Subject: "Purchase Intent Update",
		Body:    body,
		IsHTML:  true,
	}
	if err := ps.emailService.Send(email); err != nil {
		slog.Error("Error sending purchase intent update email: ", "purchase_intent_id", purchaseIntent.ID, "error", err)
	}
}

// Send payment intent notification to admins
func (ps *purchaseIntentService) PaymentIntentNotificationEmail(purchaseIntent types.PurchaseIntent) {
	admins, err := ps.serviceUser.GetAllAdmins(context.Background())
	if err != nil {
		slog.Error("Error fetching admin users: ", "error", err)
		return
	}

	// Extract emails
	adminEmails := make([]string, len(admins))
	for i, admin := range admins {
		adminEmails[i] = admin.Email
	}

	detailsLink := fmt.Sprintf("%s/admin/purchase-intents/%s", ps.baseURL, purchaseIntent.ID)
	data := map[string]string{
		"ID":          purchaseIntent.ID,
		"CustomerID":  purchaseIntent.UserID,
		"DetailsLink": detailsLink,
	}
	body, err := ps.templateService.RenderToString(PurchaseIntentNotification, data)
	if err != nil {
		slog.Error("Error loading email template: ", "error", err)
		return
	}
	email := &types.Email{
		To:      adminEmails,
		Subject: "Purchase Intent Notification",
		Body:    body,
		IsHTML:  true,
	}
	if err := ps.emailService.Send(email); err != nil {
		slog.Error("Error sending purchase intent notification email: ", "purchase_intent_id", purchaseIntent.ID, "error", err)
	}
}
