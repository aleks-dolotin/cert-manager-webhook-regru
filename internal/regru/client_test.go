package regru

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTXT_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/zone/add_txt" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("username") != "testuser" {
			t.Errorf("username = %q", r.FormValue("username"))
		}
		if r.FormValue("subdomain") != "_acme-challenge" {
			t.Errorf("subdomain = %q", r.FormValue("subdomain"))
		}
		if r.FormValue("text") != "challenge-token-123" {
			t.Errorf("text = %q", r.FormValue("text"))
		}

		resp := apiResponse{Result: "success"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("testuser", "testpass")
	c.SetBaseURL(srv.URL)

	err := c.CreateTXT("dolotin.ru", "_acme-challenge", "challenge-token-123")
	if err != nil {
		t.Fatalf("CreateTXT failed: %v", err)
	}
}

func TestDeleteTXT_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/zone/remove_record" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("record_type") != "TXT" {
			t.Errorf("record_type = %q", r.FormValue("record_type"))
		}
		if r.FormValue("content") != "challenge-token-123" {
			t.Errorf("content = %q", r.FormValue("content"))
		}

		resp := apiResponse{Result: "success"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("testuser", "testpass")
	c.SetBaseURL(srv.URL)

	err := c.DeleteTXT("dolotin.ru", "_acme-challenge", "challenge-token-123")
	if err != nil {
		t.Fatalf("DeleteTXT failed: %v", err)
	}
}

func TestCreateTXT_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := apiResponse{
			Result:    "error",
			ErrorCode: "INVALID_CREDENTIALS",
			ErrorText: "Bad username or password",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("bad", "creds")
	c.SetBaseURL(srv.URL)

	err := c.CreateTXT("dolotin.ru", "_acme-challenge", "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
