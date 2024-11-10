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
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Validate that the item was added
	addedCart, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(addedCart.Items), "Expected one item in the cart")
	assert.Equal(t, item.ProductID, addedCart.Items[0].ProductID, "Expected the same product ID")
	assert.Equal(t, item.Quantity, addedCart.Items[0].Quantity, "Expected the same quantity")
	assert.Equal(t, item.UnitPrice.Amount, addedCart.Items[0].UnitPrice.Amount, "Expected the same unit price")

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

func TestReserveCartItems_Success(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Set up product and inventory
	productID, _ := generateUUID()
	_, _ = dbPool.ExecContext(ctx, `
			INSERT INTO products (id, name, price, description)
			VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	_, _ = dbPool.ExecContext(ctx, `
			INSERT INTO inventory (product_id, quantity)
			VALUES ($1, 5)`,
		productID)

	// Create a cart for the user
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Add item to the user's cart
	item := &models.CartItem{
		ProductID: productID,
		Quantity:  1,
		UnitPrice: models.Currency{Amount: 1000},
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Test reserve_cart_items function
	err = repo.ReserveCartItems(ctx, user.ID)
	assert.NoError(t, err)

	// Cleanup
	dbPool.ExecContext(ctx, "DELETE FROM inventory_reservations WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestReserveCartItems_EmptyCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Create an empty cart for the user
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Test reserve_cart_items function for empty cart
	err = repo.ReserveCartItems(ctx, user.ID)
	assert.Error(t, err, "Expected an error for an empty cart")
	assert.Equal(t, "empty_cart", err.Error(), "Expected 'empty_cart' error message")

	// Cleanup
	dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestReserveCartItems_InsufficientInventory(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Set up product with a temporary inventory to allow adding to cart
	productID, _ := generateUUID()
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description)
		VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 1)`, // Temporarily set quantity to 1 for adding to cart
		productID)

	// Create a cart for the user
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Add item to the user's cart
	item := &models.CartItem{
		ProductID: productID,
		Quantity:  1,
		UnitPrice: models.Currency{Amount: 1000},
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Now set inventory to 0 to simulate insufficient inventory for reservation
	_, _ = dbPool.ExecContext(ctx, `
		UPDATE inventory
		SET quantity = 0
		WHERE product_id = $1`, productID)

	// Test reserve_cart_items function for insufficient inventory
	err = repo.ReserveCartItems(ctx, user.ID)
	assert.Error(t, err, "Expected an error due to insufficient inventory")
	assert.Equal(t, "insufficient_inventory", err.Error(), "Expected 'insufficient_inventory' error message")

	// Cleanup
	dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestFetchCartTotal(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Create a cart for the user
	_, err := repo.GetOrCreateCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Set up test products with known prices and quantities
	productID1, _ := generateUUID()
	productID2, _ := generateUUID()

	// Insert two products into the products table
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description)
		VALUES ($1, 'Test Product 1', 1000, 'Description 1'),
		       ($2, 'Test Product 2', 2000, 'Description 2')`,
		productID1, productID2)

	// Add inventory records for each product
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10),
		       ($2, 10)`,
		productID1, productID2)

	// Add items to the user's cart
	item1 := &models.CartItem{
		ProductID: productID1,
		Quantity:  2,
		UnitPrice: models.Currency{Amount: 1000},
	}
	item2 := &models.CartItem{
		ProductID: productID2,
		Quantity:  3,
		UnitPrice: models.Currency{Amount: 2000},
	}

	err = repo.AddItemToCart(ctx, user.ID, item1)
	assert.NoError(t, err, "Expected no error on adding item1 to cart")

	err = repo.AddItemToCart(ctx, user.ID, item2)
	assert.NoError(t, err, "Expected no error on adding item2 to cart")

	// Calculate the expected total
	expectedTotal := item1.UnitPrice.Amount*int64(item1.Quantity) +
		item2.UnitPrice.Amount*int64(item2.Quantity)

	// Fetch the cart total
	total, err := repo.FetchCartTotal(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart total")
	assert.Equal(t, expectedTotal, total.Amount, "Expected the calculated total to match")

	// Cleanup
	dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM carts WHERE user_id = $1", user.ID)
	dbPool.ExecContext(ctx, "DELETE FROM inventory WHERE product_id IN ($1, $2)", productID1, productID2)
	dbPool.ExecContext(ctx, "DELETE FROM products WHERE id IN ($1, $2)", productID1, productID2)
	dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}
