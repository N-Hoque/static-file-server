package main

import (
	"log"

	"github.com/N-Hoque/static-file-server/pkg/cli"
)

func main() {
	if err := cli.Execute(); nil != err {
		log.Fatalf("Error: %v\n", err)
	}
}
