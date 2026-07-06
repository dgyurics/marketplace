package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type OfferService interface {
	CreateOffer(ctx context.Context, offer *types.Offer) error
	UpdateOffer(ctx context.Context, offer *types.Offer) error
	GetOfferByID(ctx context.Context, id string) (types.Offer, error)
	GetOfferByIDAndUser(ctx context.Context, id string) (types.Offer, error)
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

	// notify user that offer has been received
	go ps.notificationService.NotifyOffer(offer.UserID, SubjectOfferConf, NotifyOfferConf, *offer)

	// notify admin(s) that an offer has been received
	admins, _ := ps.userService.GetAllAdmins(context.Background())
	for _, admin := range admins {
		go ps.notificationService.NotifyOffer(admin.ID, SubjectOfferRecv, NotifyOfferRecv, *offer)
	}

	return nil
}

func (ps *offerService) UpdateOffer(ctx context.Context, offer *types.Offer) error {
	if err := ps.repoOffer.UpdateOffer(ctx, offer); err != nil {
		return err
	}

	go ps.notificationService.NotifyOffer(offer.UserID, SubjectOfferUpdate, NotifyOfferUpdate, *offer)

	return nil
}

func (ps *offerService) GetOffersByProductID(ctx context.Context, id string) ([]types.Offer, error) {
	return ps.repoOffer.GetOffersByProductIDAndUser(ctx, id, getUserID(ctx))
}

func (ps *offerService) GetOfferByIDAndUser(ctx context.Context, id string) (types.Offer, error) {
	return ps.repoOffer.GetOfferByIDAndUser(ctx, id, getUserID(ctx))
}

func (ps *offerService) GetOfferByID(ctx context.Context, id string) (types.Offer, error) {
	return ps.repoOffer.GetOfferByID(ctx, id)
}

func (ps *offerService) GetOffers(ctx context.Context) ([]types.Offer, error) {
	return ps.repoOffer.GetOffers(ctx)
}
