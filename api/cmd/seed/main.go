package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ddddami/laivan/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal(fmt.Errorf("load config: %w", err))
	}

	if err := run(context.Background(), cfg.DatabaseURL, cfg.DBMaxConns); err != nil {
		fatal(err)
	}

	fmt.Println("seeded local dev marketplace data")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
	os.Exit(1)
}
