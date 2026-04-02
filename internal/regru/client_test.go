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

		var inputData map[string]interface{}
		if err := json.Unmarshal([]byte(r.FormValue("input_data")), &inputData); err != nil {
			t.Fatalf("parsing input_data: %v", err)
		}

		if inputData["username"] != "testuser" {
			t.Errorf("username = %v", inputData["username"])
		}
		if inputData["subdomain"] != "_acme-challenge" {
			t.Errorf("subdomain = %v", inputData["subdomain"])
		}
		if inputData["text"] != "challenge-token-123" {
			t.Errorf("text = %v", inputData["text"])
		}

		resp := apiResponse{Result: "success"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
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

		var inputData map[string]interface{}
		if err := json.Unmarshal([]byte(r.FormValue("input_data")), &inputData); err != nil {
			t.Fatalf("parsing input_data: %v", err)
		}

		if inputData["record_type"] != "TXT" {
			t.Errorf("record_type = %v", inputData["record_type"])
		}
		if inputData["content"] != "challenge-token-123" {
			t.Errorf("content = %v", inputData["content"])
		}

		resp := apiResponse{Result: "success"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
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
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("bad", "creds")
	c.SetBaseURL(srv.URL)

	err := c.CreateTXT("dolotin.ru", "_acme-challenge", "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
