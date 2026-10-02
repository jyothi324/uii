package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// GeocodeAddress resolves a street address to coordinates using OpenStreetMap's
// free Nominatim geocoder — no API key required. Mirrors inventory's
// utils/geocode.go so both services follow Nominatim's usage policy
// (~1 req/sec, descriptive User-Agent) identically.
const geocodeUserAgent = "TulipCRM-LeadServices/1.0 (contact: your-real-email@yourdomain.com)"

var (
	geocodeMu   sync.Mutex
	lastCallAt  time.Time
	minInterval = 1100 * time.Millisecond
)

type nominatimResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

func GeocodeAddress(street, city, state, zip, country string) (float64, float64, error) {
	geocodeMu.Lock()
	if wait := minInterval - time.Since(lastCallAt); wait > 0 {
		time.Sleep(wait)
	}
	lastCallAt = time.Now()
	geocodeMu.Unlock()

	query := fmt.Sprintf("%s, %s, %s %s, %s", street, city, state, zip, country)
	endpoint := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" + url.QueryEscape(query)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", geocodeUserAgent)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var results []nominatimResult
	if err := json.Unmarshal(body, &results); err != nil {
		return 0, 0, fmt.Errorf("unexpected geocoder response: %w", err)
	}
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no match found for address: %s", query)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, err
	}
	lng, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, err
	}

	return lat, lng, nil
}