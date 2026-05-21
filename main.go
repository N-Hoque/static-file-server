package main

import (
	"log/slog"
	"os"

	"github.com/N-Hoque/static-file-server/pkg/cli"
)

func main() {
	if err := cli.Execute(); nil != err {
		slog.Error("failed to start static file server", "error", err)
		os.Exit(1)
	}
}
