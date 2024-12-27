package repositories

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

var (
	ourEpoch   int64 = 1672531200000 // 2023-01-01T00:00:00Z in milliseconds
	seqID      int64 = 0
	seqIDMutex sync.Mutex
	shardID    int64 = 0 // Customize if using multiple instances
)

func genID() string {
	seqIDMutex.Lock()
	defer seqIDMutex.Unlock()

	// Increment sequence ID and wrap around at 1024
	seqID = (seqID + 1) % 1024

	// Get current time in milliseconds
	nowMillis := time.Now().UnixMilli()

	// Construct the ID
	result := (nowMillis - ourEpoch) << 23 // 41 bits for timestamp
	result |= (shardID << 10)              // 13 bits for shard ID
	result |= seqID                        // 10 bits for sequence ID

	return strconv.FormatInt(result, 10)
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

func TestGetOrCreateCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 2: Try to get or create a new cart for the test user
	cart, err := repo.GetOrCreateCart(ctx, user.ID) // Use the created test user's ID
	assert.NoError(t, err, "Expected no error on get or create cart")
	assert.NotNil(t, cart, "Expected cart to be returned")
	assert.Equal(t, user.ID, cart.UserID, "Expected the cart to belong to the test user")

	// Step 3: Try to get the cart again to ensure it doesn't create a duplicate
	existingCart, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching existing cart")
	assert.Equal(t, cart.UserID, existingCart.UserID, "Expected the same cart to be fetched")

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
	_, err := repo.GetOrCreateCart(ctx, user.ID) // Use the created test user's ID
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory (simulate an existing product)
	productID := genID()
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
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Validate that the item was added
	addedCart, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(addedCart.Items), "Expected one item in the cart")
	assert.Equal(t, item.ProductID, addedCart.Items[0].ProductID, "Expected the same product ID")
	assert.Equal(t, item.Quantity, addedCart.Items[0].Quantity, "Expected the same quantity")
	assert.Equal(t, item.UnitPrice, addedCart.Items[0].UnitPrice, "Expected the same unit price")

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
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	productID := genID()
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
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Update the item (increase quantity)
	item.Quantity = 2
	err = repo.UpdateCartItem(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on updating cart item")

	// Step 5: Validate that the item was updated
	updatedCart, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(updatedCart.Items), "Expected one item in the cart")
	assert.Equal(t, 2, updatedCart.Items[0].Quantity, "Expected the updated quantity")

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
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	productID := genID()
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
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Remove the item from the cart
	err = repo.RemoveItemFromCart(ctx, user.ID, productID)
	assert.NoError(t, err, "Expected no error on removing item from cart")

	// Step 5: Validate that the item was removed
	updatedCart, err := repo.GetOrCreateCart(ctx, user.ID)
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
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	productID := genID()
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
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Clear the cart
	err = repo.ClearCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on clearing cart")

	// Step 5: Validate that the cart is empty
	updatedCart, err := repo.GetOrCreateCart(ctx, user.ID)
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
