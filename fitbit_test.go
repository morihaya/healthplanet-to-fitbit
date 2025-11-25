package htf

import (
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestGetFitbitConfig(t *testing.T) {
	clientID := "test_client_id"
	clientSecret := "test_client_secret"
	cfg := GetFitbitConfig(clientID, clientSecret)

	if cfg.ClientID != clientID {
		t.Errorf("ClientID = %v, want %v", cfg.ClientID, clientID)
	}
	if cfg.ClientSecret != clientSecret {
		t.Errorf("ClientSecret = %v, want %v", cfg.ClientSecret, clientSecret)
	}
	if len(cfg.Scopes) != 1 || cfg.Scopes[0] != "weight" {
		t.Errorf("Scopes = %v, want ['weight']", cfg.Scopes)
	}
	if cfg.Endpoint.AuthURL != "https://www.fitbit.com/oauth2/authorize" {
		t.Errorf("AuthURL = %v", cfg.Endpoint.AuthURL)
	}
	if cfg.Endpoint.TokenURL != "https://api.fitbit.com/oauth2/token" {
		t.Errorf("TokenURL = %v", cfg.Endpoint.TokenURL)
	}
}

func TestNewFitbitAPI(t *testing.T) {
	clientID := "test_client_id"
	clientSecret := "test_client_secret"
	token := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		Expiry:       time.Now().Add(time.Hour),
	}

	api := NewFitbitAPI(clientID, clientSecret, token)

	if api == nil {
		t.Fatal("NewFitbitAPI returned nil")
	}
	if api.Client == nil {
		t.Error("api.Client is nil")
	}
	if api.TokenSource == nil {
		t.Error("api.TokenSource is nil")
	}
}
