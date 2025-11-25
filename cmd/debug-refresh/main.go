package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"healthplanet-to-fitbit/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ClientID: %s\n", cfg.HealthPlanet.ClientID)
	// Don't print secret
	fmt.Printf("RefreshToken: %s\n", cfg.HealthPlanet.RefreshToken)

	// Refresh Token
	values := url.Values{}
	values.Add("client_id", cfg.HealthPlanet.ClientID)
	values.Add("client_secret", cfg.HealthPlanet.ClientSecret)
	values.Add("redirect_uri", "https://www.healthplanet.jp/success.html")
	values.Add("grant_type", "refresh_token")
	values.Add("refresh_token", cfg.HealthPlanet.RefreshToken)

	res, err := http.Post("https://www.healthplanet.jp/oauth/token", "application/x-www-form-urlencoded", nil)
	// Note: http.Post doesn't support setting body with values directly like this for form-urlencoded without encoding the body string and setting reader.
	// Correct way:
	res, err = http.PostForm("https://www.healthplanet.jp/oauth/token", values)
	if err != nil {
		fmt.Printf("failed to refresh token: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Refresh Response Status: %s\n", res.Status)
	fmt.Printf("Refresh Response Body: %s\n", string(body))

	if res.StatusCode != 200 {
		fmt.Println("Refresh failed.")
		os.Exit(1)
	}

	// Parse new token (quick hack, assume it worked and we can just grab it if we wanted, but let's just use the one from response if possible or just use the existing one if refresh failed)
	// Actually, let's just try to call the API with the *existing* access token first, then the new one if we parsed it.
	// But for now, let's just see if refresh works.

	// Also try to call API with existing token
	fmt.Println("Attempting API call with EXISTING token...")
	callAPI(cfg.HealthPlanet.AccessToken)
}

func callAPI(token string) {
	values := url.Values{}
	values.Add("access_token", token)
	values.Add("date", "0")
	values.Add("tag", "6021")

	// Add from/to just in case
	now := time.Now()
	to := now.Format("20060102150405")
	from := now.AddDate(0, -1, 0).Format("20060102150405") // 1 month ago
	values.Add("from", from)
	values.Add("to", to)

	apiURL := fmt.Sprintf("https://www.healthplanet.jp/status/innerscan.json?%s", values.Encode())
	fmt.Printf("Request URL: %s\n", apiURL)

	res, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("API call failed: %v\n", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("API Response Status: %s\n", res.Status)
	fmt.Printf("API Response Body: %s\n", string(body))

	// Try with truncated token (after slash)
	// 1763952504098/LhKB...
	// Find slash
	for i, c := range token {
		if c == '/' {
			truncated := token[i+1:]
			fmt.Println("Attempting API call with TRUNCATED token...")
			callAPI(truncated)
			break
		}
	}

	// Try XML
	fmt.Println("Attempting API call with XML...")
	callAPIXML(token)
}

func callAPIXML(token string) {
	values := url.Values{}
	values.Add("access_token", token)
	values.Add("date", "1")
	values.Add("tag", "6021")

	now := time.Now()
	to := now.Format("20060102150405")
	from := now.AddDate(0, -1, 0).Format("20060102150405")
	values.Add("from", from)
	values.Add("to", to)

	apiURL := fmt.Sprintf("https://www.healthplanet.jp/status/innerscan.xml?%s", values.Encode())
	fmt.Printf("Request URL (XML): %s\n", apiURL)

	res, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("API call failed: %v\n", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("API Response Status (XML): %s\n", res.Status)
	fmt.Printf("API Response Body (XML): %s\n", string(body))

	// Try oauth_token parameter
	fmt.Println("Attempting API call with oauth_token parameter...")
	callAPIOauthToken(token)
}

func callAPIOauthToken(token string) {
	values := url.Values{}
	values.Add("oauth_token", token)
	values.Add("date", "1")
	values.Add("tag", "6021")

	now := time.Now()
	to := now.Format("20060102150405")
	from := now.AddDate(0, -1, 0).Format("20060102150405")
	values.Add("from", from)
	values.Add("to", to)

	apiURL := fmt.Sprintf("https://www.healthplanet.jp/status/innerscan.json?%s", values.Encode())
	fmt.Printf("Request URL (oauth_token): %s\n", apiURL)

	res, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("API call failed: %v\n", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("API Response Status (oauth_token): %s\n", res.Status)
	fmt.Printf("API Response Body (oauth_token): %s\n", string(body))
}
