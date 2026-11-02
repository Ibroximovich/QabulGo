package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTelegramClient_GetMe(t *testing.T) {
	mockTransport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := APIResponse[User]{
			Ok: true,
			Result: User{
				ID:        1234567,
				IsBot:     true,
				FirstName: "QabulGo Clinic Bot",
				Username:  "QabulGoClinicBot",
			},
		}
		data, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
			Header:     make(http.Header),
		}, nil
	})

	httpClient := &http.Client{Transport: mockTransport}
	client := NewClientWithHTTP("test_token", httpClient, "https://api.telegram.org")

	user, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "QabulGoClinicBot" {
		t.Errorf("expected username QabulGoClinicBot, got %s", user.Username)
	}
}

func TestTelegramClient_SendMessage_ErrorHandling(t *testing.T) {
	mockTransport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp := APIResponse[Message]{
			Ok:          false,
			ErrorCode:   403,
			Description: "Forbidden: bot was blocked by the user",
		}
		data, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(bytes.NewReader(data)),
			Header:     make(http.Header),
		}, nil
	})

	httpClient := &http.Client{Transport: mockTransport}
	client := NewClientWithHTTP("test_token", httpClient, "https://api.telegram.org")

	_, err := client.SendMessage(context.Background(), 12345, "Salom", nil)
	if err != ErrBotBlockedByUser {
		t.Errorf("expected ErrBotBlockedByUser, got %v", err)
	}
}
