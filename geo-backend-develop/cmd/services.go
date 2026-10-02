package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	baseURL      string
	trackingBase = strings.TrimRight(os.Getenv("TRACKING_BASE_URL"), "/")
)

var vehicleProfileMap = map[string]string{
	"bike":       "cycling-regular",
	"bicycle":    "cycling-regular",
	"motorcycle": "driving-car",
	"car":        "driving-car",
	"truck":      "driving-hgv",
}

var defaultProfile = "driving-car"

var smsSendCache = struct {
	sync.Mutex
	m map[string]int64
}{m: make(map[string]int64)}

func generateID(prefix string) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 3958.8
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func calculateDistanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	return calculateDistance(lat1, lng1, lat2, lng2) * 1609.344
}

func routeRecalcMinMeters() float64 {
	if v := os.Getenv("ROUTE_RECALC_MIN_METERS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			return f
		}
	}
	return 500
}

func routeRecalcMinInterval() time.Duration {
	if v := os.Getenv("ROUTE_RECALC_MIN_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			return time.Duration(i) * time.Second
		}
	}
	return 5 * time.Minute
}

func throttledDirectionsEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MAPBOX_THROTTLED_DIRECTIONS")))
	return v != "false" && v != "0"
}

func formatETA(totalHours float64) string {
	totalMinutes := int(math.Round(totalHours * 60.0))
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours == 0 && minutes == 0 {
		return "Arriving now"
	}
	if hours == 0 {
		return fmt.Sprintf("%d min", minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh", hours)
}

func formatHours(hours float64) string {
	totalMinutes := int(math.Round(hours * 60.0))
	if totalMinutes < 60 {
		return fmt.Sprintf("%d min", totalMinutes)
	}
	hoursInt := totalMinutes / 60
	minutes := totalMinutes % 60
	if minutes > 0 && hoursInt < 10 {
		return fmt.Sprintf("%dh %dm", hoursInt, minutes)
	}
	return fmt.Sprintf("%dh", hoursInt)
}

func calculateETA(distance float64) string {
	if distance > 620 {
		hours := distance / 62.0
		return formatHours(hours)
	} else if distance > 62 {
		hours := distance / 37.0
		return formatHours(hours)
	} else {
		hours := distance / 18.0
		return formatHours(hours)
	}
}

func safePhone(p string) string {
	if len(p) <= 4 {
		return "***"
	}
	return "***" + p[len(p)-4:]
}

func getPublicBaseURL(c *gin.Context) string {
	if base := strings.TrimSpace(os.Getenv("TRACKING_BASE_URL")); base != "" {
		return strings.TrimRight(base, "/")
	}
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

func buildTrackingLink(c *gin.Context, taskID string) string {
	tb := strings.TrimSpace(trackingBase)
	if tb != "" {
		if strings.Contains(tb, "{id}") {
			return strings.ReplaceAll(tb, "{id}", url.QueryEscape(taskID))
		}
		if strings.Contains(tb, ":id") {
			return strings.ReplaceAll(tb, ":id", url.QueryEscape(taskID))
		}
		if strings.Contains(tb, "%s") {
			return fmt.Sprintf(tb, url.QueryEscape(taskID))
		}
		if strings.Contains(tb, "?") {
			if strings.HasSuffix(tb, "=") {
				return tb + url.QueryEscape(taskID)
			}
			sep := "&"
			if strings.HasSuffix(tb, "?") {
				sep = ""
			}
			return fmt.Sprintf("%s%sid=%s", tb, sep, url.QueryEscape(taskID))
		}
		return fmt.Sprintf("%s?id=%s", tb, url.QueryEscape(taskID))
	}
	return fmt.Sprintf("%s/?id=%s", getPublicBaseURL(c), url.QueryEscape(taskID))
}

func getMapboxToken() string {
	return strings.TrimSpace(os.Getenv("MAPBOX_TOKEN"))
}

func loadTwilioConfig() TwilioConfig {
	return TwilioConfig{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		From:       os.Getenv("TWILIO_FROM"),
	}
}

func (tc TwilioConfig) IsConfigured() bool {
	return tc.AccountSID != "" && tc.AuthToken != "" && tc.From != ""
}

func (tc TwilioConfig) SendSMS(to string, body string) error {
	if !tc.IsConfigured() {
		return fmt.Errorf("twilio not configured: set TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_FROM")
	}
	if !strings.HasPrefix(to, "+") {
		return fmt.Errorf("phone must be in E.164 format (e.g., +15551234567)")
	}

	form := url.Values{}
	form.Set("To", to)
	form.Set("From", tc.From)
	form.Set("Body", body)

	reqURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", tc.AccountSID)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(tc.AccountSID, tc.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("twilio call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("twilio error: %s: %s", resp.Status, string(bodyBytes))
}

func trackingSMSBody(trackingURL, orderID string) string {
	if orderID != "" {
		return fmt.Sprintf("Order %s is on the way. Track live here: %s", orderID, trackingURL)
	}
	return fmt.Sprintf("Your delivery is on the way. Track live here: %s", trackingURL)
}

func smsKey(phone, trackingURLOrOrderKey, orderID string) string {
	return fmt.Sprintf("%s|%s|%s", phone, trackingURLOrOrderKey, orderID)
}

func markSmsSent(key string) {
	smsSendCache.Lock()
	defer smsSendCache.Unlock()
	smsSendCache.m[key] = time.Now().Unix()
}

func wasSmsSentRecently(key string, within time.Duration) bool {
	smsSendCache.Lock()
	defer smsSendCache.Unlock()
	t, ok := smsSendCache.m[key]
	if !ok {
		return false
	}
	return time.Since(time.Unix(t, 0)) < within
}

func geocodeAddress(addr *Address) (*Coords, error) {
	query := fmt.Sprintf("%s, %s, %s, %s, %s",
		addr.Street, addr.City, addr.State, addr.Zip, addr.County)
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, ", ,", ",")
	query = strings.Trim(query, ", ")

	if query == "" {
		return nil, fmt.Errorf("empty address")
	}

	encodedQuery := url.QueryEscape(query)
	client := &http.Client{Timeout: 10 * time.Second}

	// Try Nominatim
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", encodedQuery)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "DeliveryTrackingApp/1.0")

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var results []NominatimResponse
		if err := json.Unmarshal(body, &results); err == nil && len(results) > 0 {
			lat, _ := strconv.ParseFloat(results[0].Lat, 64)
			lon, _ := strconv.ParseFloat(results[0].Lon, 64)
			return &Coords{
				Latitude:  lat,
				Longitude: lon,
				Address:   results[0].DisplayName,
			}, nil
		}
	}

	// Try OpenCage
	openCageKey := "4bd80b00a58043ceb6062c49c5e785f5"
	if openCageKey != "" {
		openCageURL := fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s", encodedQuery, openCageKey)
		resp, _ := client.Get(openCageURL)
		if resp != nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var oc struct {
				Results []struct {
					Geometry struct {
						Lat float64 `json:"lat"`
						Lng float64 `json:"lng"`
					} `json:"geometry"`
					Formatted string `json:"formatted"`
				} `json:"results"`
			}
			if err := json.Unmarshal(body, &oc); err == nil && len(oc.Results) > 0 {
				return &Coords{
					Latitude:  oc.Results[0].Geometry.Lat,
					Longitude: oc.Results[0].Geometry.Lng,
					Address:   oc.Results[0].Formatted,
				}, nil
			}
		}
	}

	// Try Google Maps
	googleKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleKey != "" {
		googleURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", encodedQuery, googleKey)
		resp, _ := client.Get(googleURL)
		if resp != nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var g struct {
				Results []struct {
					FormattedAddress string `json:"formatted_address"`
					Geometry         struct {
						Location struct {
							Lat float64 `json:"lat"`
							Lng float64 `json:"lng"`
						} `json:"location"`
					} `json:"geometry"`
				} `json:"results"`
			}
			if err := json.Unmarshal(body, &g); err == nil && len(g.Results) > 0 {
				return &Coords{
					Latitude:  g.Results[0].Geometry.Location.Lat,
					Longitude: g.Results[0].Geometry.Location.Lng,
					Address:   g.Results[0].FormattedAddress,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no coordinates found for address: %s", query)
}

func fetchRoute(startLat, startLng, endLat, endLng float64, profile string) ([]Coords, float64, float64, error) {
	mbProfile := "mapbox/driving"
	vt := strings.ToLower(profile)

	switch {
	case strings.Contains(vt, "cycl") || vt == "bike" || vt == "bicycle":
		mbProfile = "mapbox/cycling"
	case strings.Contains(vt, "foot") || vt == "walking":
		mbProfile = "mapbox/walking"
	default:
		mbProfile = "mapbox/driving"
	}

	coordsStr := fmt.Sprintf("%f,%f;%f,%f", startLng, startLat, endLng, endLat)
	mapboxToken := getMapboxToken()
	if mapboxToken == "" {
		return nil, 0, 0, fmt.Errorf("MAPBOX_TOKEN is not configured")
	}

	apiURL := fmt.Sprintf("https://api.mapbox.com/directions/v5/%s/%s?geometries=geojson&overview=full&access_token=%s",
		mbProfile, coordsStr, mapboxToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, 0, fmt.Errorf("Mapbox API error: %d", resp.StatusCode)
	}

	var mbResp struct {
		Routes []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
		} `json:"routes"`
		Code string `json:"code"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mbResp); err != nil {
		return nil, 0, 0, err
	}

	if len(mbResp.Routes) == 0 {
		return nil, 0, 0, fmt.Errorf("no routes found")
	}

	route := mbResp.Routes[0]
	distanceMiles := route.Distance / 1609.344
	durationHours := route.Duration / 3600.0

	coords := make([]Coords, 0, len(route.Geometry.Coordinates))
	for _, coord := range route.Geometry.Coordinates {
		if len(coord) >= 2 {
			coords = append(coords, Coords{
				Latitude:  coord[1],
				Longitude: coord[0],
			})
		}
	}

	return coords, distanceMiles, durationHours, nil
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *ipRateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	times := rl.requests[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= rl.limit {
		rl.requests[key] = filtered
		return false
	}
	rl.requests[key] = append(filtered, now)
	return true
}
