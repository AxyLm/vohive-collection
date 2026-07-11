package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iniwex5/vohive/internal/config"
)

const (
	weComTimeout  = 30 * time.Second
	weComAttempts = 3
)

type WeComChannel struct {
	webhookURL string
	client     *http.Client
}

type weComTextPayload struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

type weComResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewWeComChannel(cfg config.WeComConfig) (*WeComChannel, error) {
	webhookURL := strings.TrimSpace(cfg.WebhookURL)
	if webhookURL == "" {
		return nil, errors.New("wecom webhook_url is required")
	}
	return &WeComChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: weComTimeout},
	}, nil
}

func (c *WeComChannel) Name() string { return "wecom" }

func (c *WeComChannel) Send(text string) error {
	return c.SendWithContext(NotificationContext{Event: "notification", Text: text, Timestamp: time.Now()})
}

func (c *WeComChannel) SendWithContext(ctx NotificationContext) error {
	text := strings.TrimSpace(ctx.Text)
	if text == "" {
		return nil
	}
	body, err := marshalWeComText(text)
	if err != nil {
		return err
	}
	resp, err := c.do(body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("企业微信 webhook 返回 HTTP %d", resp.StatusCode)
	}
	var out weComResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return fmt.Errorf("解析企业微信响应失败: %w", err)
	}
	if out.ErrCode != 0 {
		return fmt.Errorf("企业微信 webhook 返回 errcode=%d errmsg=%s", out.ErrCode, strings.TrimSpace(out.ErrMsg))
	}
	return nil
}

func (c *WeComChannel) do(body []byte) (*http.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= weComAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Vohive-WeCom/1.0")

		resp, err := c.client.Do(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if attempt < weComAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return nil, lastErr
}

func marshalWeComText(text string) ([]byte, error) {
	payload := weComTextPayload{MsgType: "text"}
	payload.Text.Content = text
	return json.Marshal(payload)
}

func (c *WeComChannel) RegisterCommand(_ string, _ CommandHandler) {}
func (c *WeComChannel) Start() error                               { return nil }
func (c *WeComChannel) Close() error {
	if c != nil && c.client != nil {
		c.client.CloseIdleConnections()
	}
	return nil
}
