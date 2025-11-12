package utilities

import "testing"

// US Postal Code Tests
func TestValidatePostalCode_ValidUSCode(t *testing.T) {
	tests := []string{"12345", "12345-6789"}
	for _, code := range tests {
		err := ValidatePostalCode("US", code)
		if err != nil {
			t.Errorf("Expected valid US postal code %s to return nil, got %v", code, err)
		}
	}
}

func TestValidatePostalCode_InvalidUSCode(t *testing.T) {
	tests := []string{"1234", "123456", "ABCDE", "12345-"}
	for _, code := range tests {
		err := ValidatePostalCode("US", code)
		if err == nil {
			t.Errorf("Expected invalid US postal code %s to return error", code)
		}
	}
}

// UK Postal Code Tests
func TestValidatePostalCode_ValidUKCode(t *testing.T) {
	tests := []string{"SW1A 1AA", "M1 1AA", "B33 8TH", "W1A 0AX"}
	for _, code := range tests {
		err := ValidatePostalCode("GB", code)
		if err != nil {
			t.Errorf("Expected valid UK postal code %s to return nil, got %v", code, err)
		}
	}
}

func TestValidatePostalCode_InvalidUKCode(t *testing.T) {
	tests := []string{"12345", "ABCDEFG", "SW1", "M1 1AA 1"}
	for _, code := range tests {
		err := ValidatePostalCode("GB", code)
		if err == nil {
			t.Errorf("Expected invalid UK postal code %s to return error", code)
		}
	}
}

// Canada Postal Code Tests
func TestValidatePostalCode_ValidCanadaCode(t *testing.T) {
	tests := []string{"K1A 0A6", "K1A0A6", "H3Z 2Y7", "M5V 3L9"}
	for _, code := range tests {
		err := ValidatePostalCode("CA", code)
		if err != nil {
			t.Errorf("Expected valid Canada postal code %s to return nil, got %v", code, err)
		}
	}
}

func TestValidatePostalCode_InvalidCanadaCode(t *testing.T) {
	tests := []string{"12345", "ABCDEFG", "K1A", "K1A0A6Z"}
	for _, code := range tests {
		err := ValidatePostalCode("CA", code)
		if err == nil {
			t.Errorf("Expected invalid Canada postal code %s to return error", code)
		}
	}
}

// Germany Postal Code Tests
func TestValidatePostalCode_ValidGermanyCode(t *testing.T) {
	tests := []string{"10117", "80331", "20095"}
	for _, code := range tests {
		err := ValidatePostalCode("DE", code)
		if err != nil {
			t.Errorf("Expected valid Germany postal code %s to return nil, got %v", code, err)
		}
	}
}

func TestValidatePostalCode_InvalidGermanyCode(t *testing.T) {
	tests := []string{"1234", "123456", "ABCDE"}
	for _, code := range tests {
		err := ValidatePostalCode("DE", code)
		if err == nil {
			t.Errorf("Expected invalid Germany postal code %s to return error", code)
		}
	}
}

// Netherlands Postal Code Tests
func TestValidatePostalCode_ValidNetherlandsCode(t *testing.T) {
	tests := []string{"1234 AB", "1234AB"}
	for _, code := range tests {
		err := ValidatePostalCode("NL", code)
		if err != nil {
			t.Errorf("Expected valid Netherlands postal code %s to return nil, got %v", code, err)
		}
	}
}

// Unknown Country Tests
func TestValidatePostalCode_UnknownCountry(t *testing.T) {
	err := ValidatePostalCode("XX", "12345")
	if err != nil {
		t.Errorf("Expected unknown country with postal code to return nil, got %v", err)
	}
}

// US State Tests
func TestValidateState_ValidUSState(t *testing.T) {
	tests := []string{"CA", "NY", "TX", "FL", "WA"}
	for _, state := range tests {
		err := ValidateState("US", state)
		if err != nil {
			t.Errorf("Expected valid US state %s to return nil, got %v", state, err)
		}
	}
}

func TestValidateState_InvalidUSState(t *testing.T) {
	tests := []string{"XX", "ZZ", "123", "California"}
	for _, state := range tests {
		err := ValidateState("US", state)
		if err == nil {
			t.Errorf("Expected invalid US state %s to return error", state)
		}
	}
}

// Canada Province Tests
func TestValidateState_ValidCanadaProvince(t *testing.T) {
	tests := []string{"ON", "QC", "BC", "AB", "MB"}
	for _, province := range tests {
		err := ValidateState("CA", province)
		if err != nil {
			t.Errorf("Expected valid Canada province %s to return nil, got %v", province, err)
		}
	}
}

func TestValidateState_InvalidCanadaProvince(t *testing.T) {
	tests := []string{"XX", "ZZ", "123", "Ontario"}
	for _, province := range tests {
		err := ValidateState("CA", province)
		if err == nil {
			t.Errorf("Expected invalid Canada province %s to return error", province)
		}
	}
}

// Country Without States
func TestValidateState_CountryWithoutStates(t *testing.T) {
	err := ValidateState("GB", "anystate")
	if err != nil {
		t.Errorf("Expected GB (no required states) to return nil, got %v", err)
	}
}

// Unknown Country Tests
func TestValidateState_UnknownCountry(t *testing.T) {
	err := ValidateState("XX", "anystate")
	if err != nil {
		t.Errorf("Expected unknown country state validation to return nil, got %v", err)
	}
}

// Edge Cases
func TestValidatePostalCode_EmptyCode(t *testing.T) {
	err := ValidatePostalCode("US", "")
	if err == nil {
		t.Error("Expected empty postal code to return error")
	}
}

func TestValidateState_EmptyState(t *testing.T) {
	err := ValidateState("US", "")
	if err == nil {
		t.Error("Expected empty state to return error for US")
	}
}

func TestValidatePostalCode_WhitespaceOnly(t *testing.T) {
	err := ValidatePostalCode("US", "   ")
	if err == nil {
		t.Error("Expected whitespace-only postal code to return error")
	}
}

// Case Sensitivity Tests
func TestValidatePostalCode_CaseSensitivity(t *testing.T) {
	// UK postcodes are case sensitive - lowercase should fail
	err := ValidatePostalCode("GB", "sw1a 1aa")
	if err == nil {
		t.Error("Expected lowercase UK postal code to return error (case sensitive)")
	}

	// Canada postcodes accept both cases - lowercase should work
	err = ValidatePostalCode("CA", "k1a 0a6")
	if err != nil {
		t.Errorf("Expected lowercase Canada postal code to return nil (case insensitive), got %v", err)
	}

	// Test that uppercase works for both
	err = ValidatePostalCode("GB", "SW1A 1AA")
	if err != nil {
		t.Errorf("Expected uppercase UK postal code to return nil, got %v", err)
	}

	err = ValidatePostalCode("CA", "K1A 0A6")
	if err != nil {
		t.Errorf("Expected uppercase Canada postal code to return nil, got %v", err)
	}
}

func TestValidateState_CaseSensitivity(t *testing.T) {
	// US states should be case sensitive
	err := ValidateState("US", "ca")
	if err == nil {
		t.Error("Expected lowercase US state to return error (case sensitive)")
	}

	// Canada provinces should be case sensitive
	err = ValidateState("CA", "on")
	if err == nil {
		t.Error("Expected lowercase Canada province to return error (case sensitive)")
	}
}

// Additional comprehensive tests
func TestValidatePostalCode_AdditionalCountries(t *testing.T) {
	// Test Japan
	err := ValidatePostalCode("JP", "123-4567")
	if err != nil {
		t.Errorf("Expected valid Japan postal code to return nil, got %v", err)
	}

	// Test France
	err = ValidatePostalCode("FR", "75001")
	if err != nil {
		t.Errorf("Expected valid France postal code to return nil, got %v", err)
	}

	// Test Australia
	err = ValidatePostalCode("AU", "2000")
	if err != nil {
		t.Errorf("Expected valid Australia postal code to return nil, got %v", err)
	}

	// Test Brazil
	err = ValidatePostalCode("BR", "12345-678")
	if err != nil {
		t.Errorf("Expected valid Brazil postal code to return nil, got %v", err)
	}

	// Test invalid codes for these countries
	err = ValidatePostalCode("JP", "invalid")
	if err == nil {
		t.Error("Expected invalid Japan postal code to return error")
	}

	err = ValidatePostalCode("FR", "invalid")
	if err == nil {
		t.Error("Expected invalid France postal code to return error")
	}
}
