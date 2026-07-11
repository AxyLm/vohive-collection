package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iniwex5/vohive/internal/config"
	"github.com/iniwex5/vohive/internal/notify"
)

type testWeComRequest struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
}

type testWeComResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (s *Server) handleTestWeComNotification(c *gin.Context) {
	var req testWeComRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}
	if !req.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请先启用企业微信后再测试"})
		return
	}
	webhookURL := strings.TrimSpace(req.WebhookURL)
	if webhookURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "必须填写企业微信 webhook_url"})
		return
	}

	ch, err := notify.NewWeComChannel(config.WeComConfig{Enabled: true, WebhookURL: webhookURL})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "初始化企业微信测试发送器失败: " + err.Error()})
		return
	}
	defer ch.Close()

	err = ch.SendWithContext(notify.NotificationContext{
		Event:      "wecom_test",
		Text:       "这是一条企业微信测试通知",
		DeviceID:   "test_device_001",
		DeviceName: "测试设备",
		Timestamp:  time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusOK, testWeComResponse{OK: false, Message: "测试通知发送失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, testWeComResponse{OK: true, Message: "测试通知已发送"})
}
