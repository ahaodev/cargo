package internal

import "testing"

func TestA8020033(t *testing.T) {
	result := IsClearQRString("A8020033")
	if !result {
		t.Errorf("Expected true, got false")
	}
}
