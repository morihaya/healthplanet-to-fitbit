package main

import (
	"context"
	"fmt"
	htf "healthplanet-to-fitbit"
	"healthplanet-to-fitbit/config"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	// Load environment variables
	_ = godotenv.Load(".env")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Fallback to env vars if config is empty (for backward compatibility or initial setup)
	if cfg.HealthPlanet.AccessToken == "" {
		cfg.HealthPlanet.AccessToken = os.Getenv("HEALTHPLANET_ACCESS_TOKEN")
	}
	if cfg.Fitbit.ClientID == "" {
		cfg.Fitbit.ClientID = os.Getenv("FITBIT_CLIENT_ID")
	}
	if cfg.Fitbit.ClientSecret == "" {
		cfg.Fitbit.ClientSecret = os.Getenv("FITBIT_CLIENT_SECRET")
	}
	if cfg.Fitbit.AccessToken == "" {
		cfg.Fitbit.AccessToken = os.Getenv("FITBIT_ACCESS_TOKEN")
	}
	if cfg.Fitbit.RefreshToken == "" {
		cfg.Fitbit.RefreshToken = os.Getenv("FITBIT_REFRESH_TOKEN")
	}

	// Initialize API clients
	healthPlanetAPI := htf.HealthPlanetAPI{
		AccessToken: cfg.HealthPlanet.AccessToken,
	}

	fitbitToken := &oauth2.Token{
		AccessToken:  cfg.Fitbit.AccessToken,
		RefreshToken: cfg.Fitbit.RefreshToken,
	}
	fitbitApi := htf.NewFitbitAPI(cfg.Fitbit.ClientID, cfg.Fitbit.ClientSecret, fitbitToken)

	// Parse CLI flags
	var from, to string
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 < len(args) {
				from = args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		}
	}

	// Load cache
	cacheData, err := config.LoadCache()
	if err != nil {
		log.Printf("failed to load cache: %v", err)
	}

	// Initialize Context
	ctx := context.Background()

	// Get data from HealthPlanet
	// Format dates for API (YYYYMMDDHHMM)
	var apiFrom, apiTo string
	if from != "" {
		apiFrom = from + "000000"
		apiFrom = strings.ReplaceAll(apiFrom, "-", "")
	}
	if to != "" {
		apiTo = to + "235959"
		apiTo = strings.ReplaceAll(apiTo, "-", "")
	}

	scanData, err := healthPlanetAPI.AggregateInnerScanData(ctx, apiFrom, apiTo)
	if err != nil {
		log.Fatalf("failed to aggregate inner scan data: %+v", err)
	}

	// Save data to Fitbit
	for t, data := range scanData {
		dateStr := t.Format("2006-01-02")
		if cacheData.Has(dateStr) {
			log.Printf("%s: skipped from cache", t)
			continue
		}

		weightLog, err := fitbitApi.GetBodyWeightLog(t)
		if err != nil {
			log.Printf("failed to get weight log from fitbit: %+v", err)
			break
		}

		if len(weightLog.Weight) > 0 {
			log.Printf("%s: record is found", t)
			cacheData.Add(dateStr)
			continue
		}

		if data.Weight != nil {
			if err := fitbitApi.CreateWeightLog(*data.Weight, t); err != nil {
				log.Printf("failed to create weight log: time: %s, err: %+v", t, err)
				break
			}
		}

		if data.Fat != nil {
			if err := fitbitApi.CreateBodyFatLog(*data.Fat, t); err != nil {
				log.Printf("failed to create fat log: time: %s, err: %+v", t, err)
				break
			}
		}

		printFloat := func(f *float64) string {
			if f == nil {
				return "nil"
			}
			return fmt.Sprintf("%.2f", *f)
		}

		log.Printf("%s: saved, weight: %s, fat: %s", t, printFloat(data.Weight), printFloat(data.Fat))
		cacheData.Add(dateStr)
	}

	// Save cache
	if err := config.SaveCache(cacheData); err != nil {
		log.Printf("failed to save cache: %v", err)
	}

	// Check and save token if refreshed
	newToken, err := fitbitApi.TokenSource.Token()
	if err != nil {
		log.Printf("failed to get current token: %v", err)
	} else {
		if newToken.AccessToken != cfg.Fitbit.AccessToken || newToken.RefreshToken != cfg.Fitbit.RefreshToken {
			cfg.Fitbit.AccessToken = newToken.AccessToken
			cfg.Fitbit.RefreshToken = newToken.RefreshToken
			if err := config.SaveConfig(cfg); err != nil {
				log.Printf("failed to save config: %v", err)
			} else {
				log.Printf("token refreshed and saved to config")
			}
		}
	}

	log.Printf("done")
}
