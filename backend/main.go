package main

import (
	"github.com/krystal/krystal-network-tools/backend/cmd"
	"log/slog"
	"os"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
}

func main() {
	cmd.Execute()
}
