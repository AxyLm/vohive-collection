package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

func TestWeComChannelRetriesTemporaryRequestFailure(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("response writer cannot hijack")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack: %v", err)
			}
			_ = conn.Close()
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	ch, err := NewWeComChannel(config.WeComConfig{Enabled: true, WebhookURL: srv.URL})
	if err != nil {
		t.Fatalf("NewWeComChannel() error = %v", err)
	}
	if err := ch.Send("测试"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls=%d, want 3", got)
	}
}
