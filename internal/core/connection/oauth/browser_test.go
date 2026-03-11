package oauth

import (
	"context"
	"testing"
)

func TestOpenBrowser_EmptyURL(t *testing.T) {
	t.Parallel()

	err := OpenBrowser(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
	if !containsStr(err.Error(), "URL must not be empty") {
		t.Errorf("error %q should contain 'URL must not be empty'", err.Error())
	}
}
