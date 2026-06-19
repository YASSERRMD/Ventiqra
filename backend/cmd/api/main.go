// Package main is the entry point for the Ventiqra API service.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ventiqra-api: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("ventiqra-api starting")
	return nil
}
