# AI Agent Documentation for healthplanet-to-fitbit

## Project Overview
This project synchronizes weight and body fat percentage data from [HealthPlanet](https://www.healthplanet.jp/) to Fitbit. It is written in Go.

## Directory Structure
- `cmd/`: Contains the main applications.
    - `healthplanet-to-fitbit/`: The main synchronization tool.
    - `healthplanet-gettoken/`: Tool to obtain HealthPlanet OAuth tokens.
    - `fitbit-gettoken/`: Tool to obtain Fitbit OAuth tokens.
- `config/`: Configuration handling.
- `healthplanet.go`: Client library for HealthPlanet API.
- `fitbit.go`: Client library for Fitbit API.
- `bin/`: Binary output directory.

## Key Components

### HealthPlanet API (`healthplanet.go`)
- **Purpose**: Fetches weight and body fat data from HealthPlanet.
- **Key Types**:
    - `HealthPlanetAPI`: Main client struct.
    - `InnerScanData`: Represents a single data point.
    - `AggregatedInnerScanDataMap`: Map of time to aggregated data.
- **Key Functions**:
    - `AggregateInnerScanData`: Fetches and aggregates data for a given date range. Handles 3-month chunking automatically.
    - `GetInnerScan`: Low-level API call to fetch data.

### Fitbit API (`fitbit.go`)
- **Purpose**: Uploads weight and body fat data to Fitbit.
- **Key Types**:
    - `FitbitAPI`: Main client struct using `oauth2`.
- **Key Functions**:
    - `CreateWeightLog`: Posts weight data to Fitbit.
    - `CreateBodyFatLog`: Posts body fat data to Fitbit.
    - `GetBodyWeightLog`: Retrieves existing weight logs (useful for checking existence).

## Setup and Usage

### Environment Variables
The following environment variables are required (can be in `.env`):
- `HEALTHPLANET_CLIENT_ID`
- `HEALTHPLANET_CLIENT_SECRET`
- `FITBIT_CLIENT_ID`
- `FITBIT_CLIENT_SECRET`

### Authentication
1. Run `go run cmd/healthplanet-gettoken/main.go` to get HealthPlanet tokens.
2. Run `go run cmd/fitbit-gettoken/main.go` to get Fitbit tokens.
Tokens are saved to `~/.config/healthplanet-to-fitbit/config.json`.

### Execution
- **Sync Default (last 3 months)**:
  ```bash
  go run cmd/healthplanet-to-fitbit/main.go
  ```
- **Sync Specific Range**:
  ```bash
  go run cmd/healthplanet-to-fitbit/main.go --from 2025-01-01 --to 2025-01-31
  ```

## Development Guidelines
- **Language**: Go
- **Dependency Management**: Go Modules (`go.mod`, `go.sum`)
- **Formatting**: Standard `gofmt`.
- **Error Handling**: Uses `github.com/pkg/errors` for wrapping errors.
