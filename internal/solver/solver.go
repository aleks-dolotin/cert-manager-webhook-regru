// Package solver implements the cert-manager ACME DNS-01 webhook solver
// for Reg.ru DNS.
package solver

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aleks-dolotin/cert-manager-webhook-regru/internal/regru"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
)

// RegraSolver implements webhook.Solver for Reg.ru DNS.
type RegraSolver struct {
	client *regru.Client
}

// Config is the decoded solver config from the ClusterIssuer/Issuer resource.
type Config struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

var _ webhook.Solver = &RegraSolver{}

// Name returns the solver name used in ClusterIssuer/Issuer config.
func (s *RegraSolver) Name() string {
	return "regru"
}

// Present creates a TXT record for the ACME DNS-01 challenge.
func (s *RegraSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	client, err := s.clientFromChallenge(ch)
	if err != nil {
		return err
	}

	zone := extractZone(ch.ResolvedZone)
	subdomain := extractSubdomain(ch.ResolvedFQDN, ch.ResolvedZone)

	return client.CreateTXT(zone, subdomain, ch.Key)
}

// CleanUp removes the TXT record after the challenge is completed.
func (s *RegraSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	client, err := s.clientFromChallenge(ch)
	if err != nil {
		return err
	}

	zone := extractZone(ch.ResolvedZone)
	subdomain := extractSubdomain(ch.ResolvedFQDN, ch.ResolvedZone)

	return client.DeleteTXT(zone, subdomain, ch.Key)
}

// Initialize is called when the webhook first starts.
func (s *RegraSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	// Try to create a default client from environment variables.
	// Per-challenge config can override this in clientFromChallenge.
	username := os.Getenv("REGRU_USERNAME")
	password := os.Getenv("REGRU_PASSWORD")
	if username != "" && password != "" {
		s.client = regru.NewClient(username, password)
	}
	return nil
}

// clientFromChallenge returns a Reg.ru client, preferring per-challenge config
// over the default client initialized from env vars.
func (s *RegraSolver) clientFromChallenge(ch *v1alpha1.ChallengeRequest) (*regru.Client, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return nil, err
	}

	// If config provides credentials, use them.
	if cfg.Username != "" && cfg.Password != "" {
		return regru.NewClient(cfg.Username, cfg.Password), nil
	}

	// Fall back to default client from env vars.
	if s.client != nil {
		return s.client, nil
	}

	return nil, fmt.Errorf("regru: no credentials — set REGRU_USERNAME/REGRU_PASSWORD env vars or provide in solver config")
}

func loadConfig(cfgJSON *extapi.JSON) (Config, error) {
	var cfg Config
	if cfgJSON == nil || len(cfgJSON.Raw) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("regru: decoding solver config: %w", err)
	}
	return cfg, nil
}

// extractZone removes the trailing dot from a DNS zone.
// "dolotin.ru." → "dolotin.ru"
func extractZone(zone string) string {
	return strings.TrimSuffix(zone, ".")
}

// extractSubdomain extracts the subdomain part from the FQDN relative to the zone.
// "_acme-challenge.dolotin.ru." with zone "dolotin.ru." → "_acme-challenge"
// "_acme-challenge.sub.dolotin.ru." with zone "dolotin.ru." → "_acme-challenge.sub"
func extractSubdomain(fqdn, zone string) string {
	fqdn = strings.TrimSuffix(fqdn, ".")
	zone = strings.TrimSuffix(zone, ".")
	sub := strings.TrimSuffix(fqdn, "."+zone)
	return sub
}

// New creates a new RegraSolver.
func New() *RegraSolver {
	return &RegraSolver{}
}
