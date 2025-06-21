package email

import (
	"context"
	"sparrow_blog_server/pkg/config"
	"testing"
)

func init() {
	config.LoadConfig()
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
