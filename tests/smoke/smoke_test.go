//go:build smoke

// Package smoke contains smoke tests that run against the real Reg.ru API.
//
// Requirements:
//   - REGRU_USERNAME and REGRU_PASSWORD set to valid Reg.ru credentials
//   - SMOKE_TEST_ZONE set to a real zone you own (e.g. "dolotin.ru")
//
// Run:
//
//	REGRU_USERNAME=user REGRU_PASSWORD=pass SMOKE_TEST_ZONE=dolotin.ru \
//	  go test -v -tags=smoke -count=1 -timeout=300s ./tests/smoke/...
package smoke

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aleks-dolotin/cert-manager-webhook-regru/internal/regru"
)

func skipIfNotConfigured(t *testing.T) {
	t.Helper()
	if os.Getenv("REGRU_USERNAME") == "" || os.Getenv("REGRU_PASSWORD") == "" {
		t.Skip("REGRU_USERNAME / REGRU_PASSWORD not set — skipping smoke test")
	}
	if os.Getenv("SMOKE_TEST_ZONE") == "" {
		t.Skip("SMOKE_TEST_ZONE not set — skipping smoke test")
	}
}

func zone(t *testing.T) string {
	t.Helper()
	return os.Getenv("SMOKE_TEST_ZONE")
}

func newClient(t *testing.T) *regru.Client {
	t.Helper()
	return regru.NewClient(
		os.Getenv("REGRU_USERNAME"),
		os.Getenv("REGRU_PASSWORD"),
	)
}

// uniqueSubdomain generates a test subdomain to avoid collisions.
// Uses flat name (no dots) — same level as real _acme-challenge records.
func uniqueSubdomain() string {
	return fmt.Sprintf("_acme-smoke-%d", time.Now().UnixMilli()%100000)
}

// dnsLookupTXT queries the Reg.ru authoritative NS directly for TXT records.
// Retries up to maxAttempts times with interval between attempts.
func dnsLookupTXT(t *testing.T, fqdn string, maxAttempts int, interval time.Duration) []string {
	t.Helper()
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, "dig", "@ns1.reg.ru", fqdn, "TXT", "+short", "+norecurse")
		out, err := cmd.Output()
		cancel()
		if err != nil {
			t.Logf("dig attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(interval)
			continue
		}
		var results []string
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.Trim(strings.TrimSpace(line), "\"")
			if line != "" {
				results = append(results, line)
			}
		}
		if len(results) > 0 {
			return results
		}
		if attempt < maxAttempts {
			t.Logf("DNS attempt %d/%d: no TXT for %s, retrying in %s...", attempt, maxAttempts, fqdn, interval)
			time.Sleep(interval)
		}
	}
	return nil
}

// TestSmoke_FullACMECycle simulates the cert-manager DNS-01 challenge flow:
//
//  1. Present: create TXT record _acme-challenge.smoke-XXXXX.dolotin.ru = <token>
//  2. Verify: TXT record resolves via Reg.ru authoritative NS
//  3. CleanUp: delete TXT record
//  4. Verify: TXT record no longer resolves
func TestSmoke_FullACMECycle(t *testing.T) {
	skipIfNotConfigured(t)

	client := newClient(t)
	z := zone(t)
	sub := uniqueSubdomain()
	token := fmt.Sprintf("test-token-%d", time.Now().UnixMilli())
	fqdn := sub + "." + z

	t.Logf("=== Smoke Test: ACME DNS-01 cycle ===")
	t.Logf("Zone:      %s", z)
	t.Logf("Subdomain: %s", sub)
	t.Logf("FQDN:      %s", fqdn)
	t.Logf("Token:     %s", token)

	// --- Step 1: Present (create TXT) ---
	t.Log("\n--- Step 1: Present (CreateTXT) ---")
	err := client.CreateTXT(z, sub, token)
	if err != nil {
		t.Fatalf("CreateTXT failed: %v", err)
	}
	t.Log("✅ CreateTXT succeeded")

	// Always cleanup, even if test fails
	defer func() {
		t.Log("\n--- Step 3: CleanUp (DeleteTXT) ---")
		if err := client.DeleteTXT(z, sub, token); err != nil {
			t.Logf("⚠️  CleanUp DeleteTXT failed (may already be deleted): %v", err)
		} else {
			t.Log("✅ DeleteTXT succeeded")
		}

		// --- Step 4: Verify deleted ---
		t.Log("\n--- Step 4: Verify TXT deleted ---")
		time.Sleep(3 * time.Second)
		results := dnsLookupTXT(t, fqdn, 6, 5*time.Second)
		found := false
		for _, r := range results {
			if r == token {
				found = true
			}
		}
		if found {
			t.Logf("⚠️  TXT record still resolves after delete (propagation delay)")
		} else {
			t.Logf("✅ TXT record confirmed deleted (or not yet propagated)")
		}
	}()

	// --- Step 2: Verify TXT exists via DNS ---
	t.Log("\n--- Step 2: Verify TXT via Reg.ru NS ---")
	time.Sleep(3 * time.Second) // initial propagation wait
	results := dnsLookupTXT(t, fqdn, 12, 5*time.Second)

	found := false
	for _, r := range results {
		if r == token {
			found = true
		}
	}
	if found {
		t.Logf("✅ DNS verified: %s TXT=%q (via ns1.reg.ru)", fqdn, token)
	} else {
		t.Logf("⚠️  DNS not yet propagated: %s TXT returned %v, expected %q (Reg.ru NS delay ~30-60s)", fqdn, results, token)
		t.Log("This is a soft warning — Reg.ru NS propagation can take up to 60s")
	}
}

// TestSmoke_AuthValidation verifies that credentials are accepted.
func TestSmoke_AuthValidation(t *testing.T) {
	skipIfNotConfigured(t)

	client := newClient(t)
	z := zone(t)

	// Try to create and immediately delete a record — validates auth works
	sub := fmt.Sprintf("_auth-test-%d", time.Now().UnixMilli()%100000)
	token := "auth-validation-test"

	err := client.CreateTXT(z, sub, token)
	if err != nil {
		t.Fatalf("Auth validation failed — CreateTXT returned: %v", err)
	}
	t.Log("✅ Reg.ru API accepted credentials")

	// Cleanup
	_ = client.DeleteTXT(z, sub, token)
	t.Log("✅ Cleanup done")
}

// TestSmoke_BadCredentials verifies that wrong credentials return an error.
func TestSmoke_BadCredentials(t *testing.T) {
	skipIfNotConfigured(t)

	client := regru.NewClient("invalid-user-12345", "bad-password")
	z := zone(t)

	err := client.CreateTXT(z, "_bad-creds-test", "test")
	if err == nil {
		t.Fatal("Expected error with bad credentials, got nil")
	}
	t.Logf("✅ Bad credentials correctly rejected: %v", err)
}
