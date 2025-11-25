package htf

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

// MockHTTPClient is a mock implementation of http.Client
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid network calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestInnerScanData_Time(t *testing.T) {
	tests := []struct {
		name    string
		data    InnerScanData
		want    time.Time
		wantErr bool
	}{
		{
			name: "Valid date",
			data: InnerScanData{
				Date: "202301011200",
			},
			want:    time.Date(2023, 1, 1, 12, 0, 0, 0, tz).UTC(),
			wantErr: false,
		},
		{
			name: "Invalid date",
			data: InnerScanData{
				Date: "invalid",
			},
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.data.Time()
			if (err != nil) != tt.wantErr {
				t.Errorf("InnerScanData.Time() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("InnerScanData.Time() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthPlanetAPI_AggregateInnerScanData(t *testing.T) {
	// Mock response for weight
	weightResp := `{
		"birth_date": "19900101",
		"height": "170",
		"sex": "male",
		"data": [
			{"date": "202301011200", "keydata": "70.5", "model": "test", "tag": "6021"}
		]
	}`
	// Mock response for fat
	fatResp := `{
		"birth_date": "19900101",
		"height": "170",
		"sex": "male",
		"data": [
			{"date": "202301011200", "keydata": "20.5", "model": "test", "tag": "6022"}
		]
	}`

	client := NewTestClient(func(req *http.Request) *http.Response {
		var body string
		if req.URL.Query().Get("tag") == "6021" {
			body = weightResp
		} else {
			body = fatResp
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
	})

	api := &HealthPlanetAPI{
		AccessToken: "test_token",
		Client:      client,
	}

	ctx := context.Background()
	// Test default range (empty from/to)
	got, err := api.AggregateInnerScanData(ctx, "", "")
	if err != nil {
		t.Fatalf("AggregateInnerScanData() error = %v", err)
	}

	if len(got) != 1 {
		t.Errorf("AggregateInnerScanData() got %d items, want 1", len(got))
	}

	// Check aggregated data
	targetTime := time.Date(2023, 1, 1, 12, 0, 0, 0, tz).UTC()
	data, ok := got[targetTime]
	if !ok {
		t.Fatalf("AggregateInnerScanData() data for %v not found", targetTime)
	}

	if *data.Weight != 70.5 {
		t.Errorf("Weight = %v, want 70.5", *data.Weight)
	}
	if *data.Fat != 20.5 {
		t.Errorf("Fat = %v, want 20.5", *data.Fat)
	}
}
