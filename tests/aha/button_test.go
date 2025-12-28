package aha

import (
	"testing"
)

func TestButton(t *testing.T) {
	if skipBtn {
		t.Skip("Button tests disabled")
	}

	t.Run("GetLastPressTime", ButtonGetLastPressTime)
	t.Run("String", ButtonString)
	t.Run("Buttons", ButtonButtons)
}

func ButtonGetLastPressTime(t *testing.T) {
	if len(btn.Buttons) == 0 {
		t.Skip("No buttons found on device")
	}

	for i, button := range btn.Buttons {
		pressTime := button.GetLastPressTime()
		if pressTime.Unix() == 0 {
			t.Logf("Button %d: never pressed", i)
		} else {
			t.Logf("Button %d: last pressed at %s", i, pressTime)
		}
	}
}

func ButtonString(t *testing.T) {
	s := btn.String()
	if s == "" {
		t.Fatal("ButtonDevice.String() returned empty")
	}
	t.Logf("ButtonDevice: %s", s)
}

func ButtonButtons(t *testing.T) {
	t.Logf("Device has %d buttons", len(btn.Buttons))

	for i, button := range btn.Buttons {
		s := button.String()
		if s == "" {
			t.Errorf("Button %d String() returned empty", i)
		}
	}
}
