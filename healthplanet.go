package htf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

var tz *time.Location

func init() {
	t, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}
	tz = t
}

type InnerScanTag int64

const (
	InnerScanTagWeight     InnerScanTag = 6021
	InnerScanTagBodyFatPct InnerScanTag = 6022
)

type InnerScanData struct {
	Date    string `json:"date"`
	KeyData string `json:"keydata"`
	Model   string `json:"model"`
	Tag     string `json:"tag"`
}

type AggregatedInnerScanData struct {
	Weight *float64
	Fat    *float64
}

type AggregatedInnerScanDataMap map[time.Time]*AggregatedInnerScanData

func (d *InnerScanData) Time() (time.Time, error) {
	layout := "200601021504"
	t, err := time.ParseInLocation(layout, d.Date, tz)
	if err != nil {
		return time.Time{}, err
	}

	return t.UTC(), nil
}

type InnerScanResponse struct {
	BirthDate string          `json:"birth_date"`
	Height    string          `json:"height"`
	Sex       string          `json:"sex"`
	Data      []InnerScanData `json:"data"`
}

type HealthPlanetAPI struct {
	AccessToken string
}

func (api *HealthPlanetAPI) AggregateInnerScanData(ctx context.Context, from, to string) (AggregatedInnerScanDataMap, error) {
	var weights InnerScanResponse
	var fats InnerScanResponse

	if from == "" {
		// Default behavior (last 3 months)
		var err error
		weights, err = api.GetInnerScan(ctx, InnerScanTagWeight, "", "")
		if err != nil {
			return nil, err
		}
		fats, err = api.GetInnerScan(ctx, InnerScanTagBodyFatPct, "", "")
		if err != nil {
			return nil, err
		}
	} else {
		// Parse dates
		layout := "20060102150405"
		startTime, err := time.Parse(layout, from)
		if err != nil {
			return nil, errors.Wrap(err, "invalid from date format")
		}
		endTime := time.Now()
		if to != "" {
			endTime, err = time.Parse(layout, to)
			if err != nil {
				return nil, errors.Wrap(err, "invalid to date format")
			}
		}

		// Iterate in 3-month chunks
		for current := startTime; current.Before(endTime); {
			next := current.AddDate(0, 3, 0)
			if next.After(endTime) {
				next = endTime
			}

			chunkFrom := current.Format(layout)
			chunkTo := next.Format(layout)

			w, err := api.GetInnerScan(ctx, InnerScanTagWeight, chunkFrom, chunkTo)
			if err != nil {
				return nil, err
			}
			weights.Data = append(weights.Data, w.Data...)

			f, err := api.GetInnerScan(ctx, InnerScanTagBodyFatPct, chunkFrom, chunkTo)
			if err != nil {
				return nil, err
			}
			fats.Data = append(fats.Data, f.Data...)

			current = next.Add(time.Second) // Avoid overlap
		}
	}

	m := make(AggregatedInnerScanDataMap, len(weights.Data))

	for _, weight := range weights.Data {
		t, err := weight.Time()
		if err != nil {
			log.Printf("invalid time: %+v", err)
			continue
		}

		data, err := strconv.ParseFloat(weight.KeyData, 64)
		if err != nil {
			log.Printf("invalid weight: %+v", err)
			continue
		}

		m[t] = &AggregatedInnerScanData{
			Weight: &data,
		}
	}

	for _, fat := range fats.Data {
		t, err := fat.Time()
		if err != nil {
			log.Printf("invalid time: %+v", err)
			continue
		}

		data, err := strconv.ParseFloat(fat.KeyData, 64)
		if err != nil {
			log.Printf("invalid fat: %+v", err)
			continue
		}

		if d, ok := m[t]; ok {
			d.Fat = &data
		} else {
			log.Printf("weight data not found: %+v", fat)
		}
	}

	return m, nil
}

func (api *HealthPlanetAPI) GetInnerScan(ctx context.Context, tag InnerScanTag, from, to string) (InnerScanResponse, error) {
	values := url.Values{}
	values.Add("access_token", api.AccessToken)
	values.Add("date", "1")
	if from != "" {
		values.Set("date", "1")
		values.Add("from", from)
	}
	if to != "" {
		values.Add("to", to)
	}
	values.Add("tag", strconv.Itoa(int(tag)))

	// Debug logging
	log.Printf("Requesting HealthPlanet: from=%s, to=%s, tag=%d", from, to, tag)
	// Mask token for logging
	maskedToken := "..."
	if len(api.AccessToken) > 5 {
		maskedToken = api.AccessToken[:5] + "..."
	}
	log.Printf("URL: https://www.healthplanet.jp/status/innerscan.json?access_token=%s&date=%s&from=%s&to=%s&tag=%d", maskedToken, values.Get("date"), from, to, tag)

	url := fmt.Sprintf("https://www.healthplanet.jp/status/innerscan.json?%s", values.Encode())
	res, err := http.Get(url)
	if err != nil {
		return InnerScanResponse{}, errors.Wrap(err, "failed to fetch response")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || 400 <= res.StatusCode {
		bodyBytes, _ := io.ReadAll(res.Body)
		return InnerScanResponse{}, errors.Errorf("failed to get inner scan(invalid status code): %d, body: %s", res.StatusCode, string(bodyBytes))
	}

	dec := json.NewDecoder(res.Body)
	var resData InnerScanResponse
	if err = dec.Decode(&resData); err != nil {
		return InnerScanResponse{}, errors.Wrap(err, "failed to parse response")
	}

	return resData, nil
}
