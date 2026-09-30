package ai

import "testing"

func TestNPCFallbackPersonaTemplates(t *testing.T) {
	cases := []struct {
		persona string
		want    string
	}{
		{"guard", "Hold the line!"},
		{"scout", "Eyes on the pass—they're moving east."},
		{"strategist", "Wait for the signal, then strike together."},
		{"merchant", "Keep the convoy moving; we cannot linger."},
		{"cao_cao", "Strike where they least expect it."},
		{"Cao Cao", "Strike where they least expect it."},
		{"zhang_fei", "Who dares cross this bridge?"},
		{"unknown_role", "For the realm!"},
		{"", "For the realm!"},
	}
	for _, tc := range cases {
		if got := NPCFallback(tc.persona); got != tc.want {
			t.Fatalf("persona %q: got %q want %q", tc.persona, got, tc.want)
		}
	}
}

func TestNormalizePersona(t *testing.T) {
	if normalizePersona(" Zhang Fei ") != "zhang_fei" {
		t.Fatal("normalize failed")
	}
}
