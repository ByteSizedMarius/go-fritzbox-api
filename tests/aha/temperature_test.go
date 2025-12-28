package aha

import (
	"testing"
)

func TestTemperature(t *testing.T) {
	if skipTmp {
		t.Skip("Temperature tests disabled")
	}

	t.Run("GetCelsiusNumeric", DECTGetCelsiusNumeric)
	t.Run("GetOffsetNumeric", GetOffsetNumeric)
	t.Run("GetCelsiusNumericCached", GetCelsiusNumeric)
}

func DECTGetCelsiusNumeric(t *testing.T) {
	temperature, err := tmp.DECTGetCelsiusNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature == 0 {
		t.Fatal("Reading failed: temperature is 0")
	}
}

func GetOffsetNumeric(t *testing.T) {
	offset := tmp.GetOffsetNumeric()
	t.Logf("Offset: %.1f", offset)
}

func GetCelsiusNumeric(t *testing.T) {
	temperature := tmp.GetCelsiusNumeric()
	if temperature == 0 {
		t.Fatal("Reading failed: temperature is 0")
	}
	temperature2, err := tmp.DECTGetCelsiusNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature != temperature2 {
		t.Fatalf("Temperature readings do not match: %f != %f", temperature, temperature2)
	}
}
