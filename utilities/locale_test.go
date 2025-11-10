package utilities

import "testing"

func TestValidatePostalCode_ValidUSCode(t *testing.T) {
	result := ValidatePostalCode("US", "12345")
	if !result {
		t.Error("Expected valid US postal code to return true")
	}
}

func TestValidatePostalCode_InvalidUSCode(t *testing.T) {
	result := ValidatePostalCode("US", "1234")
	if result {
		t.Error("Expected invalid US postal code to return false")
	}
}

func TestValidatePostalCode_UnknownCountry(t *testing.T) {
	result := ValidatePostalCode("XX", "12345")
	if !result {
		t.Error("Expected unknown country with postal code to return true")
	}
}

func TestValidatePostalCode_ValidUKCode(t *testing.T) {
	result := ValidatePostalCode("GB", "SW1A 1AA")
	if !result {
		t.Error("Expected valid UK postal code to return true")
	}
}

func TestValidatePostalCode_InvalidUKCode(t *testing.T) {
	result := ValidatePostalCode("GB", "12345")
	if result {
		t.Error("Expected invalid UK postal code to return false")
	}
}

func TestValidatePostalCode_ValidCanadaCode(t *testing.T) {
	result := ValidatePostalCode("CA", "K1A 0A6")
	if !result {
		t.Error("Expected valid Canada postal code to return true")
	}
}

func TestValidatePostalCode_InvalidCanadaCode(t *testing.T) {
	result := ValidatePostalCode("CA", "12345")
	if result {
		t.Error("Expected invalid Canada postal code to return false")
	}
}

func TestValidateState_ValidUSState(t *testing.T) {
	result := ValidateState("US", "CA")
	if !result {
		t.Error("Expected valid US state to return true")
	}
}

func TestValidateState_InvalidUSState(t *testing.T) {
	result := ValidateState("US", "XX")
	if result {
		t.Error("Expected invalid US state to return false")
	}
}

func TestValidateState_UnknownCountry(t *testing.T) {
	result := ValidateState("XX", "anystate")
	if !result {
		t.Error("Expected unknown country state validation to return true")
	}
}
