package models

import (
	"testing"
)

func TestProductInitialization(t *testing.T) {
	price := NewCurrency(19, 99)
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

	expectedPrice := NewCurrency(19, 99)
	if product.Price != expectedPrice {
		t.Errorf("Expected Price to be '%v', got '%v'", expectedPrice, product.Price)
	}

	if product.Description != "This is a test product" {
		t.Errorf("Expected Description to be 'This is a test product', got '%s'", product.Description)
	}
}

func TestProductStringRepresentation(t *testing.T) {
	price := NewCurrency(19, 99)
	product := Product{
		Price: price,
	}

	expectedPriceString := "$19.99"
	if product.Price.String() != expectedPriceString {
		t.Errorf("Expected Price string to be '%s', got '%s'", expectedPriceString, product.Price.String())
	}
}
