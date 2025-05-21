package services

import (
	"context"
	pingttl "github.com/strideynet/go-ping-ttl"
	"log/slog"
	"os"
)

var (
	stdPing = pingttl.New()
)

func GetPinger() *pingttl.Pinger {
	return stdPing
}

func StartPinger(ctx context.Context) {
	go func() {
		if err := stdPing.Run(ctx); err != nil {
			slog.Error("pinger error", "error", err)
			os.Exit(1)
		}
	}()
}
