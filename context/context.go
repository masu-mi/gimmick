package context

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
)

// WithCancelBySignal returns context cancelable by signal
func WithCancelBySignal(parent context.Context, sigs ...os.Signal) (ctx context.Context) {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		sc := make(chan os.Signal)
		signal.Notify(sc, sigs...)
		select {
		case <-ctx.Done():
		case <-sc:
		}
		cancel()
	}()
	return ctx
}

// NewHTTPRequest returns *http.Requst related with ctx.
func NewHTTPRequest(ctx context.Context, method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return req, err
	}
	req.WithContext(ctx)
	return req, nil
}
