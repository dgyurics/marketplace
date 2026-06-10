package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type OfferService interface {
	CreateOffer(ctx context.Context, offer *types.Offer) error
	UpdateOffer(ctx context.Context, offer *types.Offer) error
	GetOfferByID(ctx context.Context, id string) (types.Offer, error)
	GetOffersByProductID(ctx context.Context, id string) ([]types.Offer, error)
	GetOffers(ctx context.Context) ([]types.Offer, error)
}

func NewOfferService(
	repoProduct repositories.ProductRepository,
	repoOffer repositories.OfferRepository,
	userService UserService,
	productService ProductService,
	notificationService NotificationService,
) OfferService {
	return &offerService{
		repoOffer:           repoOffer,
		repoProduct:         repoProduct,
		userService:         userService,
		productService:      productService,
		notificationService: notificationService,
	}
}

type offerService struct {
	repoOffer           repositories.OfferRepository
	repoProduct         repositories.ProductRepository
	userService         UserService
	productService      ProductService
	notificationService NotificationService
}

func (ps *offerService) CreateOffer(ctx context.Context, offer *types.Offer) (err error) {
	offer.UserID = getUserID(ctx)
	offer.ID, err = utilities.GenerateIDString()
	if err != nil {
		return err
	}
	offer.Status = types.OfferPending

	if err := ps.repoOffer.CreateOffer(ctx, offer); err != nil {
		return err
	}

	go ps.OfferNotification(*offer)

	return nil
}

func (ps *offerService) UpdateOffer(ctx context.Context, offer *types.Offer) error {
	if err := ps.repoOffer.UpdateOffer(ctx, offer); err != nil {
		return err
	}

	// email customer if offer status changes to accepted or rejected
	if offer.Status == types.OfferAccepted || offer.Status == types.OfferRejected {
		product, err := ps.productService.GetProductByID(ctx, offer.Product.ID)
		if err != nil {
			return err
		}
		offer.Product = product
		go ps.OfferUpdateEmail(*offer)
	}
	return nil
}

func (ps *offerService) GetOffersByProductID(ctx context.Context, id string) ([]types.Offer, error) {
	return ps.repoOffer.GetOffersByProductIDAndUser(ctx, id, getUserID(ctx))
}

func (ps *offerService) GetOfferByID(ctx context.Context, id string) (types.Offer, error) {
	return ps.repoOffer.GetOfferByID(ctx, id)
}

func (ps *offerService) GetOffers(ctx context.Context) ([]types.Offer, error) {
	return ps.repoOffer.GetOffers(ctx)
}

func (ps *offerService) OfferUpdateEmail(offer types.Offer) {
	usr, err := ps.userService.GetUserByID(context.Background(), offer.UserID)
	if err != nil {
		slog.Error("Error fetching user", "ID", offer.UserID, "error", err)
		return
	}

	data := map[string]string{
		"ProductName": offer.Product.Name,
		"Status":      string(offer.Status),
	}
	if usr.Email == nil {
		slog.Debug("Error sending offer update email, nil email")
		return
	}
	if err := ps.notificationService.SendEmail(*usr.Email, "Offer Update", OfferUpdate, data); err != nil {
		slog.Error("Error sending offer update email: ", "offer_id", offer.ID, "error", err)
	}
}

// Offer received notification to admins
func (ps *offerService) OfferNotification(offer types.Offer) {
	admins, err := ps.userService.GetAllAdmins(context.Background())
	if err != nil {
		slog.Error("Error fetching admin users: ", "error", err)
		return
	}

	detailsLink := fmt.Sprintf("%s/admin/offers/%s", ps.notificationService.BaseURL(), offer.ID)
	data := map[string]string{
		"OfferID":     offer.ID,
		"CustomerID":  offer.UserID,
		"DetailsLink": detailsLink,
	}

	for _, admin := range admins {
		if err := ps.notificationService.Notify(admin.ID, OfferNotificationAdmin, data); err != nil {
			slog.Error("Error sending offer notification to admins: ", "offer_id", offer.ID, "error", err)
		}
	}
}
