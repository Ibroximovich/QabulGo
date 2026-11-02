package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrBotBlockedByUser = errors.New("bot foydalanuvchi tomonidan bloklangan (403 Forbidden)")
	ErrChatNotFound     = errors.New("chat topilmadi (400 Bad Request)")
	ErrTooManyRequests  = errors.New("telegram rate limit cheklovi (429 Too Many Requests)")
)

type Client interface {
	GetMe(ctx context.Context) (*User, error)
	SendMessage(ctx context.Context, chatID int64, text string, replyMarkup any) (*Message, error)
	EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, replyMarkup *InlineKeyboardMarkup) (*Message, error)
	AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error
	GetUpdates(ctx context.Context, offset int64, limit int, timeoutSeconds int) ([]Update, error)
}

type client struct {
	token      string
	apiBaseURL string
	httpClient *http.Client
}

func NewClient(token string, customBaseURL ...string) Client {
	return NewClientWithHTTP(token, &http.Client{Timeout: 45 * time.Second}, customBaseURL...)
}

func NewClientWithHTTP(token string, httpClient *http.Client, customBaseURL ...string) Client {
	baseURL := "https://api.telegram.org"
	if len(customBaseURL) > 0 && strings.TrimSpace(customBaseURL[0]) != "" {
		baseURL = strings.TrimRight(customBaseURL[0], "/")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 45 * time.Second}
	}

	return &client{
		token:      token,
		apiBaseURL: baseURL,
		httpClient: httpClient,
	}
}

func (c *client) makeRequest(ctx context.Context, method string, payload any, responseTarget any) error {
	url := fmt.Sprintf("%s/bot%s/%s", c.apiBaseURL, c.token, method)

	var bodyReader io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("telegram payload marshaling xatoligi: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return fmt.Errorf("telegram so'rovini yaratishda xatolik: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram API aloqa xatoligi: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("telegram javobini o'qishda xatolik: %w", err)
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrBotBlockedByUser
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrTooManyRequests
	}

	if err := json.Unmarshal(bodyBytes, responseTarget); err != nil {
		return fmt.Errorf("telegram javobini parslashda xatolik (HTTP %d): %w", resp.StatusCode, err)
	}

	return nil
}

func (c *client) GetMe(ctx context.Context) (*User, error) {
	var resp APIResponse[User]
	if err := c.makeRequest(ctx, "getMe", nil, &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("telegram getMe muvaffaqiyatsiz: %s", resp.Description)
	}
	return &resp.Result, nil
}

func (c *client) SendMessage(ctx context.Context, chatID int64, text string, replyMarkup any) (*Message, error) {
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	var resp APIResponse[Message]
	if err := c.makeRequest(ctx, "sendMessage", payload, &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("telegram sendMessage muvaffaqiyatsiz (kod %d): %s", resp.ErrorCode, resp.Description)
	}
	return &resp.Result, nil
}

func (c *client) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, replyMarkup *InlineKeyboardMarkup) (*Message, error) {
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	var resp APIResponse[Message]
	if err := c.makeRequest(ctx, "editMessageText", payload, &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("telegram editMessageText muvaffaqiyatsiz (kod %d): %s", resp.ErrorCode, resp.Description)
	}
	return &resp.Result, nil
}

func (c *client) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string, showAlert bool) error {
	payload := map[string]any{
		"callback_query_id": callbackQueryID,
	}
	if text != "" {
		payload["text"] = text
		payload["show_alert"] = showAlert
	}

	var resp APIResponse[bool]
	if err := c.makeRequest(ctx, "answerCallbackQuery", payload, &resp); err != nil {
		return err
	}
	return nil
}

func (c *client) GetUpdates(ctx context.Context, offset int64, limit int, timeoutSeconds int) ([]Update, error) {
	payload := map[string]any{
		"offset":          offset,
		"limit":           limit,
		"timeout":         timeoutSeconds,
		"allowed_updates": []string{"message", "callback_query"},
	}

	var resp APIResponse[[]Update]
	if err := c.makeRequest(ctx, "getUpdates", payload, &resp); err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, fmt.Errorf("telegram getUpdates muvaffaqiyatsiz: %s", resp.Description)
	}
	return resp.Result, nil
}
