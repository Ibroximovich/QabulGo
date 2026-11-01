package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"qabulgo-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_NilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CORS(nil) panik qildi: %v", r)
		}
	}()

	r := gin.New()
	r.Use(CORS(nil))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Kutilgan 200, lekin %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin kutilgan http://localhost:5173, lekin %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_EmptyAndInvalidFrontendURL(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Bo'sh yoki noto'g'ri URL bilan CORS panik qildi: %v", r)
		}
	}()

	cfg := &config.Config{
		AppEnv:      "testing",
		FrontendURL: "   , *, bad-schema, ftp://invalid.com,   ",
	}

	r := gin.New()
	r.Use(CORS(cfg))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Localhost dev origin ruxsat etilishi kerak
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Kutilgan 200, lekin %d", w.Code)
	}

	// * yoki yaroqsiz origin ruxsat etilmasligi kerak
	badReq, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	badReq.Header.Set("Origin", "http://evil.com")
	badW := httptest.NewRecorder()
	r.ServeHTTP(badW, badReq)

	if badW.Code != http.StatusForbidden {
		t.Fatalf("Yaroqsiz origin rad etilishi kerak (403), lekin %d qaytdi", badW.Code)
	}
}

func TestCORS_ProductionSecurity(t *testing.T) {
	cfg := &config.Config{
		AppEnv:      "production",
		FrontendURL: "https://qabulgo.uz",
	}

	r := gin.New()
	r.Use(CORS(cfg))
	r.OPTIONS("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 1. Ruxsat berilgan production origin
	req, _ := http.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://qabulgo.uz")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://qabulgo.uz" {
		t.Fatalf("Access-Control-Allow-Origin noto'g'ri: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("Access-Control-Allow-Credentials true bo'lishi kerak")
	}

	// 2. Production da localhost ruxsat etilmasligi kerak agar ko'rsatilmagan bo'lsa
	devReq, _ := http.NewRequest(http.MethodOptions, "/ping", nil)
	devReq.Header.Set("Origin", "http://localhost:3000")
	devReq.Header.Set("Access-Control-Request-Method", "POST")
	devW := httptest.NewRecorder()
	r.ServeHTTP(devW, devReq)

	if devW.Code != http.StatusForbidden {
		t.Fatalf("Productionda localhost rad etilishi kerak (403), lekin %d qaytdi", devW.Code)
	}
}

