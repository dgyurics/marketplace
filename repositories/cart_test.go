package repositories

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a unique test user
func createUniqueTestUser(t *testing.T, userRepo UserRepository) *types.User {
	ctx := context.Background()

	// Generate a unique email using random numbers and current timestamp
	randomSuffix := mathrand.Intn(1000000)
	email := fmt.Sprintf("testuser%d@example.com", randomSuffix)
	user := types.User{ID: utilities.MustGenerateIDString()}

	// Create a new user object
	user.Email = email
	user.PasswordHash = "hashedpassword"

	// Insert the user into the database
	err := userRepo.CreateUser(ctx, &user)
	assert.NoError(t, err, "Expected no error on user creation")
	assert.NotEmpty(t, user.ID, "Expected user ID to be set")

	return &user
}

// Helper function to create a unique guest user
func createUniqueGuestUser(t *testing.T, userRepo UserRepository) *types.User {
	ctx := context.Background()

	// Create a new guest user object
	user := types.User{ID: utilities.MustGenerateIDString()}

	// Insert the guest user into the database
	err := userRepo.CreateGuest(ctx, &user)
	assert.NoError(t, err, "Expected no error on guest user creation")
	assert.NotEmpty(t, user.ID, "Expected guest user ID to be set")

	return &user
}

func TestGetCart(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 2: Try to get or create a new cart for the test user
	cart, err := repo.GetCart(ctx, user.ID) // Use the created test user's ID
	assert.NoError(t, err, "Expected no error on get or create cart")
	assert.NotNil(t, cart, "Expected cart to be returned")

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
	_, err := repo.GetCart(ctx, user.ID) // Use the created test user's ID
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory (simulate an existing product)
	product := types.Product{ID: utilities.MustGenerateIDString()}

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary)
		VALUES ($1, 'Test Product', 1000, 'Test product summary')`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &types.CartItem{
		Product:   product,
		Quantity:  1,
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Validate that the item was added
	cart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(cart), "Expected one item in the cart")
	assert.Equal(t, item.Product.ID, cart[0].Product.ID, "Expected the same product ID")
	assert.Equal(t, item.Quantity, cart[0].Quantity, "Expected the same quantity")
	assert.Equal(t, item.UnitPrice, cart[0].UnitPrice, "Expected the same unit price")

	// Clean up the cart, product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart items")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
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
	_, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	product := types.Product{ID: utilities.MustGenerateIDString()}

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary)
		VALUES ($1, 'Test Product', 1000, 'Test product summary')`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &types.CartItem{
		Product:   product,
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
	updatedCart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(updatedCart), "Expected one item in the cart")
	assert.Equal(t, 2, updatedCart[0].Quantity, "Expected the updated quantity")

	// Clean up the cart, product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart items")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
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
	_, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	product := types.Product{ID: utilities.MustGenerateIDString()}

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary)
		VALUES ($1, 'Test Product', 1000, 'Test product summary')`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &types.CartItem{
		Product:   product,
		Quantity:  1,
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Remove the item from the cart
	err = repo.RemoveItemFromCart(ctx, user.ID, product.ID)
	assert.NoError(t, err, "Expected no error on removing item from cart")

	// Step 5: Validate that the item was removed
	updatedCart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 0, len(updatedCart), "Expected no items in the cart")

	// Clean up the product and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
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
	_, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	product := types.Product{ID: utilities.MustGenerateIDString()}

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary)
		VALUES ($1, 'Test Product', 1000, 'Test product summary')`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add an item to the cart
	item := &types.CartItem{
		Product:   product,
		Quantity:  1,
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 4: Clear the cart
	err = repo.ClearCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on clearing cart")

	// Step 5: Validate that the cart is empty
	updatedCart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 0, len(updatedCart), "Expected no items in the cart")

	// Clean up the product, and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on deleting product")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting user")
}

func TestGetCartWithImages(t *testing.T) {
	repo := NewCartRepository(dbPool)
	userRepo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// Step 1: Create a new cart for the test user
	_, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on cart creation")

	// Step 2: Add a valid product to the inventory
	product := types.Product{ID: utilities.MustGenerateIDString()}

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary)
		VALUES ($1, 'Test Product', 1000, 'Test product summary')`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting test product")

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 10)`,
		product.ID)
	assert.NoError(t, err, "Expected no error on inserting inventory")

	// Step 3: Add images to the product
	imgID1 := utilities.MustGenerateIDString()
	imgID2 := utilities.MustGenerateIDString()
	imageIDs := []string{imgID1, imgID2}
	imageURLs := []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"}
	imageSrces := []string{"image1.jpg", "image2.jpg"}

	for i, imageID := range imageIDs {
		_, err = dbPool.ExecContext(ctx, `
			INSERT INTO images (id, product_id, url, source)
			VALUES ($1, $2, $3, $4)`,
			imageID, product.ID, imageURLs[i], imageSrces[i])
		assert.NoError(t, err, "Expected no error on inserting product images")
	}

	// Step 4: Add an item to the cart
	item := &types.CartItem{
		Product:   product,
		Quantity:  1,
		UnitPrice: 100000,
	}
	err = repo.AddItemToCart(ctx, user.ID, item)
	assert.NoError(t, err, "Expected no error on adding item to cart")

	// Step 5: Retrieve the cart and verify images
	cart, err := repo.GetCart(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on fetching cart")
	assert.Equal(t, 1, len(cart), "Expected one item in the cart")

	// Verify that product images are included
	fetchedProduct := cart[0].Product
	assert.Equal(t, product.ID, fetchedProduct.ID, "Expected correct product ID")
	assert.GreaterOrEqual(t, len(fetchedProduct.Images), 2, "Expected at least 2 images for the product")
	assert.Equal(t, imageURLs[0], fetchedProduct.Images[1].URL, "Expected correct image URL")
	assert.Equal(t, imageURLs[1], fetchedProduct.Images[0].URL, "Expected correct image URL")

	// Cleanup
	_, err = dbPool.ExecContext(ctx, "DELETE FROM cart_items WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting cart items")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM images WHERE product_id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on deleting images")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on deleting product")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on deleting user")
}
