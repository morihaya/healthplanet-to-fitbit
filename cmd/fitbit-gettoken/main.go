package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	htf "healthplanet-to-fitbit"
	"healthplanet-to-fitbit/config"
)

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func genCodeChallenge() (verifier string, challenge string) {
	verifier = randomString(128)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	return
}

func main() {
	godotenv.Load(".env")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	clientID := cfg.Fitbit.ClientID
	if clientID == "" {
		clientID = os.Getenv("FITBIT_CLIENT_ID")
	}
	if clientID == "" {
		fmt.Print("Input Fitbit Client ID: ")
		fmt.Scan(&clientID)
	}

	clientSecret := cfg.Fitbit.ClientSecret
	if clientSecret == "" {
		clientSecret = os.Getenv("FITBIT_CLIENT_SECRET")
	}
	if clientSecret == "" {
		fmt.Print("Input Fitbit Client Secret: ")
		fmt.Scan(&clientSecret)
	}

	conf := htf.GetFitbitConfig(clientID, clientSecret)

	verifier, challenge := genCodeChallenge()

	server := &http.Server{Addr: ":8080"}
	done := make(chan struct{})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		ctx := context.Background()
		token, err := conf.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "err: %v", err)
			return
		}

		cfg.Fitbit.ClientID = clientID
		cfg.Fitbit.ClientSecret = clientSecret
		cfg.Fitbit.AccessToken = token.AccessToken
		cfg.Fitbit.RefreshToken = token.RefreshToken
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Fprintf(w, "failed to save config: %v", err)
			return
		}

		fmt.Fprintf(w, "AccessToken: %s\n", token.AccessToken)
		fmt.Fprintf(w, "RefreshToken: %s\n", token.RefreshToken)
		fmt.Fprintf(w, "Credentials saved to config file. You can close this window.")
		close(done)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := conf.AuthCodeURL("state",
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.AccessTypeOffline,
		)

		http.Redirect(w, r, url, http.StatusFound)
	})

	fmt.Println("Open: http://localhost:8080")
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-done
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
	}
	fmt.Println("Token saved successfully. Exiting.")
}
