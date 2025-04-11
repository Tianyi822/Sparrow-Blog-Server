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
	err := SendVerificationCodeBySys(ctx)
	if err != nil {
		t.Fatalf("SendVerificationCodeBySys failed: %v", err)
	}

	err = SendVerificationCodeBySys(ctx)
	if err != nil {
		t.Fatalf("SendVerificationCodeBySys failed: %v", err)
	}
}
