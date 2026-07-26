package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ddddami/laivan/internal/config"
	"github.com/ddddami/laivan/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal(fmt.Errorf("load config: %w", err))
	}

	ctx := context.Background()
	var mediaUploader storage.Uploader
	if cfg.Media.Enabled {
		uploader, err := storage.NewS3Uploader(ctx, cfg.Media)
		if err != nil {
			fatal(fmt.Errorf("create seed media uploader: %w", err))
		}
		mediaUploader = uploader
	}

	if err := run(ctx, cfg.DatabaseURL, cfg.DBMaxConns, mediaUploader); err != nil {
		fatal(err)
	}

	fmt.Println("seeded local dev marketplace data")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
	os.Exit(1)
}
