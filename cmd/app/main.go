package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cawa87/garantex-test/internal/app"
	"github.com/cawa87/garantex-test/internal/config"
)

func main() {
	// Parse command line flags
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and run application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	fmt.Println("Starting Garantex Rate Service...")
	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}

	os.Exit(0)
}
