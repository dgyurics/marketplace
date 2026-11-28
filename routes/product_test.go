package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	util "github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dummyAuth struct{}

func (d dummyAuth) RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
}

func TestProductMemberPriceVisibility(t *testing.T) {
	for _, tc := range []struct {
		name    string
		user    *types.User
		visible bool
	}{
		{"anonymous", nil, false},
		{"guest", &types.User{Role: types.RoleGuest}, false},
		{"user", &types.User{Role: types.RoleUser}, false},
		{"member", &types.User{Role: types.RoleMember}, true},
		{"staff", &types.User{Role: types.RoleStaff}, true},
		{"admin", &types.User{Role: types.RoleAdmin}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			product := types.Product{ID: "1", MemberPrice: util.Ptr(int64(0))}
			service := new(MockProductService)
			service.On("GetProductByID", mock.Anything, "1").Return(product, nil)
			service.On("GetProducts", mock.Anything, mock.Anything).Return([]types.Product{product}, nil)
			h := &ProductRoutes{productService: service}
			for _, list := range []bool{false, true} {
				req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
				req.SetPathValue("id", "1")
				if tc.user != nil {
					req = req.WithContext(context.WithValue(req.Context(), services.UserKey, tc.user))
				}
				rr := httptest.NewRecorder()
				if list {
					h.GetProducts(rr, req)
				} else {
					h.GetProduct(rr, req)
				}
				require.Equal(t, http.StatusOK, rr.Code)
				if tc.visible {
					require.Contains(t, rr.Body.String(), `"member_price":0`)
				} else {
					require.NotContains(t, rr.Body.String(), `"member_price"`)
				}
			}
		})
	}
}

func (d dummyAuth) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return next
}

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *types.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]types.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(ctx context.Context, id string) (types.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(types.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, product types.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) RemoveProduct(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGetProductByID(t *testing.T) {
	mockService := new(MockProductService)

	routes := &ProductRoutes{
		productService: mockService,
		router: router{
			mux:            http.NewServeMux(),
			authMiddleware: &dummyAuth{},
		},
	}

	routes.RegisterRoutes()

	product := types.Product{
		ID:          "1",
		Name:        "Test Product",
		Price:       util.Ptr(int64(100000)),
		Summary:     util.Ptr("This is a test product summary"),
		Description: util.Ptr("This is a test product description"),
		Inventory:   10,
	}

	mockService.On("GetProductByID", mock.Anything, "1").Return(product, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rr := httptest.NewRecorder()
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	expected, err := json.Marshal(product)
	require.NoError(t, err)
	require.JSONEq(t, string(expected), rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetProducts(t *testing.T) {
	mockService := new(MockProductService)

	routes := &ProductRoutes{
		productService: mockService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	products := []types.Product{
		{
			ID:          "1",
			Name:        "Test Product 1",
			Price:       util.Ptr(int64(100000)),
			Summary:     util.Ptr("This is the first test product summary"),
			Description: util.Ptr("This is the first test product description"),
		},
		{
			ID:          "2",
			Name:        "Test Product 2",
			Price:       util.Ptr(int64(200000)),
			Summary:     util.Ptr("This is the second test product summary"),
			Description: util.Ptr("This is the second test product description"),
		},
	}

	mockService.On("GetProducts", mock.Anything, mock.Anything).Return(products, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("GET /products", routes.GetProducts)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	expected, err := json.Marshal(products)
	require.NoError(t, err)
	require.JSONEq(t, string(expected), rr.Body.String())

	mockService.AssertExpectations(t)
}

func TestUpdateProductWithoutStatus(t *testing.T) {
	mockService := new(MockProductService)
	routes := &ProductRoutes{productService: mockService}
	mockService.On("UpdateProduct", mock.Anything, mock.MatchedBy(func(product types.Product) bool {
		return product.ID == "1" && product.Status == nil && product.Summary != nil && *product.Summary == "Updated summary"
	})).Return(nil).Once()

	req := httptest.NewRequest(http.MethodPut, "/products", strings.NewReader(`{"id":"1","name":"Test Product","price":1000,"summary":"Updated summary"}`))
	rr := httptest.NewRecorder()
	routes.UpdateProduct(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}
