package repositories

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mathrand "math/rand"
	"testing"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

func generateUUID() (string, error) {
	u := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, u); err != nil {
		return "", err
	}

	u[6] = (u[6] & 0x0f) | 0x40 // Set the version to 4
	u[8] = (u[8] & 0x3f) | 0x80 // Set the variant to RFC 4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:]), nil
}

// Helper function to create a unique test user
func createUniqueTestUser(t *testing.T, userRepo UserRepository) *models.User {
	ctx := context.Background()

	// Generate a unique email and phone using random numbers and current timestamp
	randomSuffix := mathrand.Intn(1000000)
	email := fmt.Sprintf("testuser%d@example.com", randomSuffix)
	phone := fmt.Sprintf("12345%06d", randomSuffix)

	// Create a new user object
	user := &models.User{
		Email:        email,
		Phone:        phone,
		PasswordHash: "hashedpassword",
	}

	// Insert the user into the database
	err := userRepo.CreateUser(ctx, user)
	assert.NoError(t, err, "Expected no error on user creation")
	assert.NotEmpty(t, user.ID, "Expected user ID to be set")

	return user
}

func TestCreateCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 2: Create a new cart for the test user
	err := repo.CreateCart(ctx, user.ID) // Use the created test user's ID
	assert.NoError(t, err, "Expected no error on cart creation")

	// Clean up the cart
	_, err = dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on cart deletion")

	// Clean up the test user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestAddItemToCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 1: Create a new cart for the test user
	err := repo.CreateCart(ctx, user.ID) // Use the created test user's ID
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory (simulate an existing product)
	productID, err := generateUUID()
	assert.NoError(t, err, "Expected no error on generating UUID")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description)
		VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &models.CartItem{
		ProductID: productID, // Use the valid UUID
		Quantity:  1,
		UnitPrice: models.Currency{Amount: 1000},
		TotalPrice: models.Currency{
			Amount: 1000,
		},
	}
	err = repo.AddItemToCart(ctx, user.ID, item) // Use user.ID instead of cart.ID
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Validate that the item was added
	addedCart, err := repo.GetCart(ctx, user.ID) // Use user.ID instead of cart.ID
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(addedCart.Items), "Expected one item in the cart")
	assert.Equal(t, item.ProductID, addedCart.Items[0].ProductID, "Expected the same product ID")

	// Clean up the cart, product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart items")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	assert.NoError(t, err, "Expected no error on deleting product")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting user")
}

func TestUpdateCartItem(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 1: Create a new cart for the test user
	err := repo.CreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	productID, err := generateUUID()
	assert.NoError(t, err, "Expected no error on generating UUID")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description)
		VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &models.CartItem{
		ProductID: productID,
		Quantity:  1,
		UnitPrice: models.Currency{Amount: 1000},
		TotalPrice: models.Currency{
			Amount: 1000,
		},
	}
	err = repo.AddItemToCart(ctx, user.ID, item) // Use user.ID instead of cart.ID
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Update the item (increase quantity)
	item.Quantity = 2
	item.TotalPrice.Amount = 2000
	err = repo.UpdateCartItem(ctx, user.ID, item) // Use user.ID instead of cart.ID
	assert.NoError(t, err, "Expected no error on updating cart item")

	// Step 5: Validate that the item was updated
	updatedCart, err := repo.GetCart(ctx, user.ID) // Use user.ID instead of cart.ID
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(updatedCart.Items), "Expected one item in the cart")
	assert.Equal(t, 2, updatedCart.Items[0].Quantity, "Expected the updated quantity")
	assert.Equal(t, int64(2000), updatedCart.Items[0].TotalPrice.Amount, "Expected the updated total price")

	// Clean up the cart, product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart items")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	assert.NoError(t, err, "Expected no error on deleting product")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting user")
}

func TestRemoveItemFromCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 1: Create a new cart for the test user
	err := repo.CreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	productID, err := generateUUID()
	assert.NoError(t, err, "Expected no error on generating UUID")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description)
		VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &models.CartItem{
		ProductID: productID,
		Quantity:  1,
		UnitPrice: models.Currency{Amount: 1000},
		TotalPrice: models.Currency{
			Amount: 1000,
		},
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Remove the item from the cart
	err = repo.RemoveItemFromCart(ctx, user.ID, productID)
	assert.NoError(t, err, "Expected no error on removing item from cart")

	// Step 5: Validate that the item was removed
	updatedCart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 0, len(updatedCart.Items), "Expected no items in the cart")

	// Clean up the cart, product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	assert.NoError(t, err, "Expected no error on deleting product")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting user")
}

func TestClearCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 1: Create a new cart for the test user
	err := repo.CreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	productID, err := generateUUID()
	assert.NoError(t, err, "Expected no error on generating UUID")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description)
		VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		productID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &models.CartItem{
		ProductID: productID,
		Quantity:  1,
		UnitPrice: models.Currency{Amount: 1000},
		TotalPrice: models.Currency{
			Amount: 1000,
		},
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Clear the cart
	err = repo.ClearCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on clearing cart")

	// Step 5: Validate that the cart is empty
	updatedCart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 0, len(updatedCart.Items), "Expected no items in the cart")

	// Clean up the cart, product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	assert.NoError(t, err, "Expected no error on deleting product")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting user")
}
