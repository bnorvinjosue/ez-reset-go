package app

import "testing"

func TestModelsLoadsFromInternalDevices(t *testing.T) {
	a := New()
	if a.dbError != nil {
		t.Fatalf("failed to load device database: %v", a.dbError)
	}
	models, err := a.Models()
	if err != nil {
		t.Fatalf("Models() error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("expected at least one supported model")
	}
}

func TestParseField(t *testing.T) {
	id := "MFG:EPSON;MDL:Stylus COLOR 760;CMD:ESCPL2;DES:EPSON Stylus COLOR 760;"
	if got := parseField(id, "MDL"); got != "Stylus COLOR 760" {
		t.Fatalf("parseField MDL = %q, want %q", got, "Stylus COLOR 760")
	}
	if got := parseField(id, "DES"); got != "EPSON Stylus COLOR 760" {
		t.Fatalf("parseField DES = %q, want %q", got, "EPSON Stylus COLOR 760")
	}
	if got := parseField(id, "NOPE"); got != "" {
		t.Fatalf("parseField NOPE = %q, want empty", got)
	}
}
