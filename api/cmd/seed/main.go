package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func main() {
	databaseURL := os.Getenv("LAIVAN_DB_URL")
	if databaseURL == "" {
		fatal(errors.New("LAIVAN_DB_URL is required"))
	}

	if err := run(context.Background(), databaseURL); err != nil {
		fatal(err)
	}

	fmt.Println("seeded local dev marketplace data")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
	os.Exit(1)
}
