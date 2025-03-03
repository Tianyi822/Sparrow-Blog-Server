package email

import (
	"context"
	"h2blog_server/pkg/config"
	"testing"
)

func init() {
	_ = config.LoadConfig()
}

func TestSendVerificationCodeEmail(t *testing.T) {
	ctx := context.Background()
	err := SendVerificationCodeEmail(ctx, "chentyit@163.com")
	if err != nil {
		t.Fatalf("SendVerificationCodeEmail failed: %v", err)
	}
}
