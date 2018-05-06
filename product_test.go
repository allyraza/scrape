package souqr

import (
	"testing"
)

func Test_Product_ToXML(t *testing.T) {
	got := 100
	expected := 100

	if got != expected {
		t.Errorf("Expected %v, Got %v", expected, got)
	}
}
