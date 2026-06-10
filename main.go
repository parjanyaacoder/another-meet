package main

import (
	"fmt"
	"os"

	"github.com/parjanyaacoder/another-meet/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		os.Exit(1)
	}
}

