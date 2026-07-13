package geo

import (
	"net"
	"testing"
)

func TestIsPrivate(t *testing.T) {
	tests := []struct {
		ip      string
		private bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}
	for _, tt := range tests {
		if got := IsPrivate(net.ParseIP(tt.ip)); got != tt.private {
			t.Errorf("IsPrivate(%s) = %v, want %v", tt.ip, got, tt.private)
		}
	}
}

func TestIsPrivate_IPv6(t *testing.T) {
	tests := []struct {
		ip      string
		private bool
	}{
		{"::1", true},               // loopback
		{"fc00::1", true},           // unique-local
		{"fe80::1", true},           // link-local
		{"2001:db8::1", false},      // documentation (not private per Go)
		{"2606:4700:4700::1111", false}, // public Cloudflare DNS
	}
	for _, tt := range tests {
		if got := IsPrivate(net.ParseIP(tt.ip)); got != tt.private {
			t.Errorf("IsPrivate(%s) = %v, want %v", tt.ip, got, tt.private)
		}
	}
}

func TestIsPrivate_InvalidIP(t *testing.T) {
	if got := IsPrivate(net.ParseIP("not-an-ip")); got {
		t.Error("expected false for invalid IP")
	}
}
