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

	// no email, just inbox notification?
	go ps.OfferConfirmation(*offer)
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
		go ps.OfferUpdate(*offer)
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

// Offer status change notification
// TODO move this out of offer service (find better place)
func (ps *offerService) OfferUpdate(offer types.Offer) {
	detailsLink := fmt.Sprintf("%s/offers/%s", ps.notificationService.BaseURL(), offer.ID)
	data := map[string]string{
		"Status":      string(offer.Status),
		"DetailsLink": detailsLink,
	}
	if err := ps.notificationService.Notify(offer.UserID, "Offer Update", NotifyOfferUpdate, data); err != nil {
		slog.Error("Error sending offer update: ", "offer_id", offer.ID, "user_id", offer.UserID, "error", err)
	}
}

// Offer confirmation notification to user
// TODO move this out of offer service (find better place)
func (ps *offerService) OfferConfirmation(offer types.Offer) {
	detailsLink := fmt.Sprintf("%s/offers/%s", ps.notificationService.BaseURL(), offer.ID)
	data := map[string]string{
		"DetailsLink": detailsLink,
	}
	if err := ps.notificationService.Notify(offer.UserID, SubjectOfferConf, NotifyOfferConf, data); err != nil {
		slog.Error("Error sending offer confirmation: ", "offer_id", offer.ID, "user_id", offer.UserID, "error", err)
	}
}

// Offer received notification to admin
// TODO move this out of offer service (find better place)
func (ps *offerService) OfferNotification(offer types.Offer) {
	admins, err := ps.userService.GetAllAdmins(context.Background())
	if err != nil {
		slog.Error("Error fetching admin users: ", "error", err)
		return
	}

	detailsLink := fmt.Sprintf("%s/admin/offers/%s", ps.notificationService.BaseURL(), offer.ID)
	data := map[string]string{
		"DetailsLink": detailsLink,
	}

	for _, admin := range admins {
		if err := ps.notificationService.Notify(admin.ID, SubjectOfferRecv, NotifyOfferRecv, data); err != nil {
			slog.Error("Error sending offer received: ", "offer_id", offer.ID, "user_id", admin.ID, "error", err)
		}
	}
}
