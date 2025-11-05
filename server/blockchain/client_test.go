package blockchain

import "testing"

func TestSanitizeTxHash(t *testing.T) {
	validUpper := "F7C7DAFC7674FF8D026973E6BE05363EFD450A92EA461EDF77B009F11ED7E417"
	expected := "f7c7dafc7674ff8d026973e6be05363efd450a92ea461edf77b009f11ed7e417"

	if got := sanitizeTxHash(validUpper); got != expected {
		t.Fatalf("sanitizeTxHash(%q) = %q, want %q", validUpper, got, expected)
	}

	if got := sanitizeTxHash("not-a-hash"); got != "" {
		t.Fatalf("expected empty string for invalid hash, got %q", got)
	}

	doubleEncoded := "46374337444146433736373446463844303236393733453642453035333633454644343530413932454134363145444637374230303946313145443745343137"
	if got := sanitizeTxHash(doubleEncoded); got != "" {
		t.Fatalf("expected sanitizer to reject double-encoded hash, got %q", got)
	}
}

func TestComputeTxHash(t *testing.T) {
	txBytes := []byte("sample tx bytes")
	expected := "69bfb45c5eb1ff6f7555a052075effe33a51ac855a53ce7345b4f032c27d4e96"

	if got := computeTxHash(txBytes); got != expected {
		t.Fatalf("computeTxHash returned %q, want %q", got, expected)
	}

	if computeTxHash(nil) != "" {
		t.Fatalf("expected empty string for nil input")
	}
}
