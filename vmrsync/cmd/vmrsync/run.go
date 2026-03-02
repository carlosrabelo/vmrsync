package main

import (
	"fmt"
	"os"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/version"
)

func run(osArgs []string) {
	if len(osArgs) >= 2 && osArgs[1] == "version" {
		fmt.Printf("VM RSync v%s\n", version.Version)
		os.Exit(0)
	}
	os.Exit(0)
}
