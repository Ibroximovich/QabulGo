package utils

import (
	"testing"
)

func TestExtractCoordinatesFromURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedLat float64
		expectedLon float64
		expectNil   bool
	}{
		{
			name:        "Google place with comma",
			url:         "https://www.google.com/maps/place/37.244245,67.277320/data=...",
			expectedLat: 37.244245,
			expectedLon: 67.277320,
			expectNil:   false,
		},
		{
			name:        "Google @lat,lon",
			url:         "https://www.google.com/maps/@37.239278,67.293741,15z",
			expectedLat: 37.239278,
			expectedLon: 67.293741,
			expectNil:   false,
		},
		{
			name:        "Google ?q=lat,lon",
			url:         "https://maps.google.com/?q=37.234278,67.292365",
			expectedLat: 37.234278,
			expectedLon: 67.292365,
			expectNil:   false,
		},
		{
			name:        "Yandex ll=lon,lat",
			url:         "https://yandex.uz/maps/?ll=67.292365,37.234278&z=16",
			expectedLat: 37.234278,
			expectedLon: 67.292365,
			expectNil:   false,
		},
		{
			name:        "Google data !3d!4d",
			url:         "https://www.google.com/maps/search/?api=1&query=test&data=!3d37.244245!4d67.277320",
			expectedLat: 37.244245,
			expectedLon: 67.277320,
			expectNil:   false,
		},
		{
			name:      "Empty string",
			url:       "",
			expectNil: true,
		},
		{
			name:      "Invalid URL without coords",
			url:       "https://google.com/search?q=hospital",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon := ExtractCoordinatesFromURL(tt.url)
			if tt.expectNil {
				if lat != nil || lon != nil {
					t.Errorf("expected nil coordinates, got lat=%v lon=%v", lat, lon)
				}
				return
			}
			if lat == nil || lon == nil {
				t.Fatalf("expected non-nil coordinates, got nil")
			}
			diffLat := *lat - tt.expectedLat
			if diffLat < -0.0001 || diffLat > 0.0001 {
				t.Errorf("lat mismatch: got %f, want %f", *lat, tt.expectedLat)
			}
			diffLon := *lon - tt.expectedLon
			if diffLon < -0.0001 || diffLon > 0.0001 {
				t.Errorf("lon mismatch: got %f, want %f", *lon, tt.expectedLon)
			}
		})
	}
}

