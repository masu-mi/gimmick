package context

import (
	"context"
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
