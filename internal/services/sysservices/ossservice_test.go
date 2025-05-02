package sysservices

import (
	"context"
	"h2blog_server/storage"
	"testing"
)

func TestGetPresignUrlById(t *testing.T) {
	ctx := context.Background()
	url, err := GetPresignUrlById(ctx, "0ab6f800e0ea3270")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("url = %v\n", url)
	storage.Storage.Close(ctx)
}
