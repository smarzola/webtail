package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// Print version if requested
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("webtail %s\n", version)
		os.Exit(0)
	}

	// Parse command-line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Loaded configuration with %d services", len(config.Services))

	// Create proxies for each service
	var proxies []*Proxy
	for _, serviceConfig := range config.Services {
		proxy := NewProxy(&serviceConfig, &config.Tailscale)
		proxies = append(proxies, proxy)
	}

	// Start all proxies
	var wg sync.WaitGroup
	cancel := func() {} // placeholder cancel function

	startedProxies := 0
	for _, proxy := range proxies {
		wg.Add(1)
		go func(p *Proxy) {
			defer wg.Done()
			if err := p.Start(); err != nil {
				log.Printf("Failed to start proxy for %s: %v", p.config.NodeName, err)
				return
			}
			log.Printf("Started proxy for %s", p.config.NodeName)
		}(proxy)
		startedProxies++
	}

	if startedProxies == 0 {
		log.Fatal("No proxies could be started")
	}

	log.Printf("Started %d proxies. Press Ctrl+C to stop.", startedProxies)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal, stopping proxies...")

	// Cancel context to signal shutdown
	cancel()

	// Stop all proxies with timeout
	done := make(chan struct{})
	go func() {
		var stopWg sync.WaitGroup
		for _, proxy := range proxies {
			stopWg.Add(1)
			go func(p *Proxy) {
				defer stopWg.Done()
				if err := p.Stop(); err != nil {
					log.Printf("Error stopping proxy for %s: %v", p.config.NodeName, err)
				}
			}(proxy)
		}
		stopWg.Wait()
		close(done)
	}()

	// Wait for graceful shutdown or timeout
	select {
	case <-done:
		log.Println("All proxies stopped gracefully")
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting for proxies to stop")
	}

	log.Println("Shutdown complete")
}
