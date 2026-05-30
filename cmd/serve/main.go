package main

import (
	"log/slog"
	"os"

	"github.com/N-Hoque/static-file-server/internal/cli"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	if err := cli.Execute(); nil != err {
		slog.Error("failed to start static file server", "error", err)
		os.Exit(1)
	}
}
