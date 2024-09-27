package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

type Currency struct {
	Amount int64 // Amount in cents
}

func NewCurrency(dollars, cents int64) Currency {
	return Currency{Amount: dollars*100 + cents}
}

func (c *Currency) Add(dollars, cents int64) {
	c.Amount += dollars*100 + cents
}

func (c Currency) String() string {
	return fmt.Sprintf("$%.2f", float64(c.Amount)/100) // %.2f follows standard rounding rules (round half up)
}

// MarshalJSON customizes the JSON representation of Currency
func (c Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.2f", float64(c.Amount)/100))
}

// UnmarshalJSON customizes the JSON parsing of Currency
func (c *Currency) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	c.Amount = int64(math.Round(parsedValue * 100))
	return nil
}

// Scan implements the sql.Scanner interface, allowing Currency to be used in Scan directly.
func (c *Currency) Scan(value interface{}) error {
	switch v := value.(type) {
	case float64:
		c.Amount = int64(math.Round(v * 100))
	case string:
		parsedValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("Currency.Scan: cannot convert string to float64: %w", err)
		}
		c.Amount = int64(math.Round(parsedValue * 100))
	case []byte:
		strValue := string(v)
		parsedValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return fmt.Errorf("Currency.Scan: cannot convert []byte to float64: %w", err)
		}
		c.Amount = int64(math.Round(parsedValue * 100))
	default:
		return fmt.Errorf("Currency.Scan: expected float64 or string but got %T", value)
	}
	return nil
}
