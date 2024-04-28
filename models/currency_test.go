package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewCurrency(t *testing.T) {
	c := NewCurrency(19, 99)
	expectedAmount := int64(1999)
	if c.Amount != expectedAmount {
		t.Errorf("NewCurrency(19, 99) = %d; want %d", c.Amount, expectedAmount)
	}
}

func TestAdd(t *testing.T) {
	c := NewCurrency(19, 99)
	c.Add(5, 50)
	expectedAmount := int64(2549)
	if c.Amount != expectedAmount {
		t.Errorf("Add(5, 50) = %d; want %d", c.Amount, expectedAmount)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		currency Currency
		expected string
	}{
		{NewCurrency(19, 99), "$19.99"},
		{NewCurrency(0, 1), "$0.01"},
		{NewCurrency(123456, 78), "$123456.78"},
		{Currency{Amount: 100}, "$1.00"},
		{Currency{Amount: 1000}, "$10.00"},
		{Currency{Amount: 1001}, "$10.01"},
	}

	for _, test := range tests {
		result := test.currency.String()
		if result != test.expected {
			t.Errorf("String() = %s; want %s", result, test.expected)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	c := NewCurrency(19, 99)
	expected := `"19.99"`

	jsonData, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("MarshalJSON() failed: %v", err)
	}

	if string(jsonData) != expected {
		t.Errorf("MarshalJSON() = %s; want %s", jsonData, expected)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	jsonData := `"19.99"`
	var c Currency

	if err := json.Unmarshal([]byte(jsonData), &c); err != nil {
		t.Fatalf("UnmarshalJSON() failed: %v", err)
	}

	expectedAmount := int64(1999)
	if c.Amount != expectedAmount {
		t.Errorf("UnmarshalJSON() = %d; want %d", c.Amount, expectedAmount)
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int64
	}{
		{float64(19.99), 1999},
		{"19.99", 1999},
	}

	for _, test := range tests {
		var c Currency
		err := c.Scan(test.input)
		if err != nil {
			t.Fatalf("Scan(%v) failed: %v", test.input, err)
		}

		if c.Amount != test.expected {
			t.Errorf("Scan(%v) = %d; want %d", test.input, c.Amount, test.expected)
		}
	}

	// Test with an invalid type
	var c Currency
	err := c.Scan([]byte{0x01})
	if err == nil {
		t.Errorf("Scan() should have failed with invalid type")
	}
}

func TestMarshalUnmarshalJSONRoundTrip(t *testing.T) {
	c := NewCurrency(19, 99)
	jsonData, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("MarshalJSON() failed: %v", err)
	}

	var result Currency
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("UnmarshalJSON() failed: %v", err)
	}

	if !reflect.DeepEqual(result, c) {
		t.Errorf("Marshal/Unmarshal round-trip = %+v; want %+v", result, c)
	}
}
