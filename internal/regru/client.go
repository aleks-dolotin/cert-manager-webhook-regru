// Package regru implements a minimal Reg.ru API v2 client
// for creating and deleting TXT DNS records.
//
// This is a standalone implementation that talks directly to the Reg.ru API
// using username/password authentication. It does NOT import code from
// external-dns-regru-webhook — only two API calls are needed (add_txt, remove_record).
package regru

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.reg.ru/api/regru2"

// Client is a minimal Reg.ru API v2 client for TXT record management.
type Client struct {
	baseURL  string
	username string
	password string
	http     *http.Client
}

// NewClient creates a new Reg.ru API client.
func NewClient(username, password string) *Client {
	return &Client{
		baseURL:  DefaultBaseURL,
		username: username,
		password: password,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// SetBaseURL overrides the API base URL (useful for testing).
func (c *Client) SetBaseURL(u string) {
	c.baseURL = strings.TrimRight(u, "/")
}

// apiResponse is the top-level Reg.ru API v2 response envelope.
type apiResponse struct {
	Result string `json:"result"`
	Answer *struct {
		Domains []domainResult `json:"domains"`
	} `json:"answer"`
	ErrorCode string `json:"error_code"`
	ErrorText string `json:"error_text"`
}

type domainResult struct {
	Result string `json:"result"`
	Dname  string `json:"dname"`
}

// CreateTXT creates a TXT record in the given zone.
func (c *Client) CreateTXT(zone, subdomain, content string) error {
	params := url.Values{
		"username":      {c.username},
		"password":      {c.password},
		"domains":       {fmt.Sprintf(`[{"dname":"%s"}]`, zone)},
		"subdomain":     {subdomain},
		"text":          {content},
		"output_format": {"json"},
	}

	return c.doRequest("/zone/add_txt", params)
}

// DeleteTXT removes a TXT record from the given zone.
func (c *Client) DeleteTXT(zone, subdomain, content string) error {
	params := url.Values{
		"username":      {c.username},
		"password":      {c.password},
		"domains":       {fmt.Sprintf(`[{"dname":"%s"}]`, zone)},
		"subdomain":     {subdomain},
		"content":       {content},
		"record_type":   {"TXT"},
		"output_format": {"json"},
	}

	return c.doRequest("/zone/remove_record", params)
}

func (c *Client) doRequest(endpoint string, params url.Values) error {
	reqURL := c.baseURL + endpoint

	resp, err := c.http.PostForm(reqURL, params)
	if err != nil {
		return fmt.Errorf("regru: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("regru: reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("regru: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("regru: parsing response: %w", err)
	}

	if apiResp.Result != "success" {
		return fmt.Errorf("regru: API error: %s (%s)", apiResp.ErrorText, apiResp.ErrorCode)
	}

	// Check per-domain results
	if apiResp.Answer != nil {
		for _, d := range apiResp.Answer.Domains {
			if d.Result != "success" {
				return fmt.Errorf("regru: domain %s error: %s", d.Dname, d.Result)
			}
		}
	}

	return nil
}
