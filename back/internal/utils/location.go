package utils

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Google: /place/37.244245,67.277320 or /place/37.244245+67.277320
	googlePlaceRegex = regexp.MustCompile(`(?i)/place/(-?\d+(?:\.\d+)?)[,+ ](-?\d+(?:\.\d+)?)`)
	// Google: @37.244245,67.277320
	googleAtRegex = regexp.MustCompile(`@(-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?)`)
	// Google: ?q=37.244245,67.277320 or &daddr=37.244245,67.277320
	googleQueryRegex = regexp.MustCompile(`(?i)[?&](?:q|daddr)=(-?\d+(?:\.\d+)?)[,+ ](-?\d+(?:\.\d+)?)`)
	// Google data: !3d37.244245!4d67.277320
	googleDataRegex = regexp.MustCompile(`(?i)!3d(-?\d+(?:\.\d+)?)(?:!4d|%214d)(-?\d+(?:\.\d+)?)`)
	// Yandex: ?ll=67.277320,37.244245 or ?pt=67.277320,37.244245 (NOTE: Yandex is lon,lat)
	yandexLLRegex = regexp.MustCompile(`(?i)[?&](?:ll|pt)=(-?\d+(?:\.\d+)?)[,%](-?\d+(?:\.\d+)?)`)
)

func isValidLat(lat float64) bool {
	return lat >= -90.0 && lat <= 90.0
}

func isValidLon(lon float64) bool {
	return lon >= -180.0 && lon <= 180.0
}

// ExtractCoordinatesFromURL parses latitude and longitude from Google / Yandex map links
func ExtractCoordinatesFromURL(rawURL string) (*float64, *float64) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, nil
	}

	// 1. Direct regex check on raw / decoded URL
	lat, lon := extractFromURLString(trimmed)
	if lat != nil && lon != nil {
		return lat, lon
	}

	// 2. Check if it's a short URL or redirecting link
	lower := strings.ToLower(trimmed)
	isShort := strings.Contains(lower, "maps.app.goo.gl") ||
		strings.Contains(lower, "goo.gl/maps") ||
		strings.Contains(lower, "yandex.ru/maps/-/") ||
		strings.Contains(lower, "yandex.uz/maps/-/") ||
		strings.Contains(lower, "bit.ly")

	if isShort {
		resolvedLat, resolvedLon := resolveShortURL(trimmed)
		if resolvedLat != nil && resolvedLon != nil {
			return resolvedLat, resolvedLon
		}
	}

	return nil, nil
}

func extractFromURLString(rawURL string) (*float64, *float64) {
	decoded, err := url.QueryUnescape(rawURL)
	if err != nil {
		decoded = rawURL
	}

	// 1. Google Place
	if m := googlePlaceRegex.FindStringSubmatch(decoded); len(m) == 3 {
		if lat, err1 := strconv.ParseFloat(m[1], 64); err1 == nil && isValidLat(lat) {
			if lon, err2 := strconv.ParseFloat(m[2], 64); err2 == nil && isValidLon(lon) {
				return &lat, &lon
			}
		}
	}

	// 2. Google At (@lat,lon)
	if m := googleAtRegex.FindStringSubmatch(decoded); len(m) == 3 {
		if lat, err1 := strconv.ParseFloat(m[1], 64); err1 == nil && isValidLat(lat) {
			if lon, err2 := strconv.ParseFloat(m[2], 64); err2 == nil && isValidLon(lon) {
				return &lat, &lon
			}
		}
	}

	// 3. Google Query (?q=lat,lon or ?daddr=lat,lon)
	if m := googleQueryRegex.FindStringSubmatch(decoded); len(m) == 3 {
		if lat, err1 := strconv.ParseFloat(m[1], 64); err1 == nil && isValidLat(lat) {
			if lon, err2 := strconv.ParseFloat(m[2], 64); err2 == nil && isValidLon(lon) {
				return &lat, &lon
			}
		}
	}

	// 4. Google Data (!3dlat!4dlon)
	if m := googleDataRegex.FindStringSubmatch(decoded); len(m) == 3 {
		if lat, err1 := strconv.ParseFloat(m[1], 64); err1 == nil && isValidLat(lat) {
			if lon, err2 := strconv.ParseFloat(m[2], 64); err2 == nil && isValidLon(lon) {
				return &lat, &lon
			}
		}
	}

	// 5. Yandex (ll=lon,lat or pt=lon,lat)
	if m := yandexLLRegex.FindStringSubmatch(decoded); len(m) == 3 {
		if lon, err1 := strconv.ParseFloat(m[1], 64); err1 == nil && isValidLon(lon) {
			if lat, err2 := strconv.ParseFloat(m[2], 64); err2 == nil && isValidLat(lat) {
				return &lat, &lon
			}
		}
	}

	return nil, nil
}

func resolveShortURL(shortURL string) (*float64, *float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var finalLat, finalLon *float64

	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 6 {
				return http.ErrUseLastResponse
			}
			redirectURL := req.URL.String()
			if lat, lon := extractFromURLString(redirectURL); lat != nil && lon != nil {
				finalLat = lat
				finalLon = lon
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, shortURL, nil)
	if err != nil {
		return nil, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if finalLat != nil && finalLon != nil {
			return finalLat, finalLon
		}
		if loc := resp.Header.Get("Location"); loc != "" {
			if lat, lon := extractFromURLString(loc); lat != nil && lon != nil {
				return lat, lon
			}
		}
	}

	// If HEAD didn't yield, try GET
	reqGet, err := http.NewRequestWithContext(ctx, http.MethodGet, shortURL, nil)
	if err != nil {
		return nil, nil
	}
	reqGet.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	respGet, err := client.Do(reqGet)
	if err == nil {
		defer respGet.Body.Close()
		if finalLat != nil && finalLon != nil {
			return finalLat, finalLon
		}
		finalURL := respGet.Request.URL.String()
		return extractFromURLString(finalURL)
	}

	return nil, nil
}

