package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iniwex5/vohive/internal/config"
)

func TestWeComChannelSendsTextPayload(t *testing.T) {
	var got weComTextPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type=%q", ct)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("payload json: %v", err)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	ch, err := NewWeComChannel(config.WeComConfig{Enabled: true, WebhookURL: srv.URL})
	if err != nil {
		t.Fatalf("NewWeComChannel() error = %v", err)
	}

	if err := ch.SendWithContext(NotificationContext{Event: "sms_received", Text: "企业微信测试通知"}); err != nil {
		t.Fatalf("SendWithContext() error = %v", err)
	}
	if got.MsgType != "text" {
		t.Fatalf("msgtype=%q", got.MsgType)
	}
	if got.Text.Content != "企业微信测试通知" {
		t.Fatalf("content=%q", got.Text.Content)
	}
}

func TestWeComChannelReportsErrCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":93000,"errmsg":"invalid webhook"}`))
	}))
	defer srv.Close()

	ch, err := NewWeComChannel(config.WeComConfig{Enabled: true, WebhookURL: srv.URL})
	if err != nil {
		t.Fatalf("NewWeComChannel() error = %v", err)
	}
	if err := ch.Send("测试"); err == nil {
		t.Fatal("expected errcode failure")
	}
}
