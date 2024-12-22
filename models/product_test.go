package models

import (
	"testing"
)

func TestProductInitialization(t *testing.T) {
	price := int64(1999)
	product := Product{
		ID:          "123",
		Name:        "Test Product",
		Price:       price,
		Description: "This is a test product",
	}

	if product.ID != "123" {
		t.Errorf("Expected ID to be '123', got '%s'", product.ID)
	}

	if product.Name != "Test Product" {
		t.Errorf("Expected Name to be 'Test Product', got '%s'", product.Name)
	}

	expectedPrice := int64(1999)
	if product.Price != expectedPrice {
		t.Errorf("Expected Price to be '%v', got '%v'", expectedPrice, product.Price)
	}

	if product.Description != "This is a test product" {
		t.Errorf("Expected Description to be 'This is a test product', got '%s'", product.Description)
	}
}
