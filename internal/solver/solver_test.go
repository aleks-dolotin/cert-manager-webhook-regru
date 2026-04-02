package solver

import "testing"

func TestExtractZone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"dolotin.ru.", "dolotin.ru"},
		{"example.com.", "example.com"},
		{"dolotin.ru", "dolotin.ru"},
	}
	for _, tc := range tests {
		got := extractZone(tc.input)
		if got != tc.want {
			t.Errorf("extractZone(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		fqdn string
		zone string
		want string
	}{
		{"_acme-challenge.dolotin.ru.", "dolotin.ru.", "_acme-challenge"},
		{"_acme-challenge.sub.dolotin.ru.", "dolotin.ru.", "_acme-challenge.sub"},
		{"_acme-challenge.example.com.", "example.com.", "_acme-challenge"},
	}
	for _, tc := range tests {
		got := extractSubdomain(tc.fqdn, tc.zone)
		if got != tc.want {
			t.Errorf("extractSubdomain(%q, %q) = %q, want %q", tc.fqdn, tc.zone, got, tc.want)
		}
	}
}

func TestName(t *testing.T) {
	s := New()
	if s.Name() != "regru" {
		t.Errorf("Name() = %q, want %q", s.Name(), "regru")
	}
}
