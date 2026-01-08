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
	dockerEnabled := flag.Bool("docker", false, "Enable Docker container discovery")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig(*configPath, *dockerEnabled)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Loaded configuration with %d services", len(config.Services))

	// Create proxies for each service from config
	var proxies []*Proxy
	for _, serviceConfig := range config.Services {
		proxy := NewProxy(&serviceConfig, &config.Tailscale)
		proxies = append(proxies, proxy)
	}

	// Start all config-based proxies
	var wg sync.WaitGroup

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

	// Start Docker watcher if enabled
	var dockerWatcher *DockerWatcher
	if *dockerEnabled {
		log.Printf("Docker discovery enabled on network %q, starting Docker watcher...", config.Docker.Network)
		dockerWatcher, err = NewDockerWatcher(&config.Tailscale, &config.Docker)
		if err != nil {
			log.Printf("Warning: Failed to create Docker watcher: %v", err)
		} else {
			if err := dockerWatcher.Start(); err != nil {
				log.Printf("Warning: Failed to start Docker watcher: %v", err)
				dockerWatcher = nil
			}
		}
	}

	if startedProxies == 0 && dockerWatcher == nil {
		log.Fatal("No proxies could be started and Docker watcher is not running")
	}

	if startedProxies > 0 {
		log.Printf("Started %d config-based proxies.", startedProxies)
	}
	if dockerWatcher != nil {
		log.Println("Docker watcher is running for dynamic container discovery.")
	}
	log.Println("Press Ctrl+C to stop.")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal, stopping...")

	// Stop Docker watcher first
	if dockerWatcher != nil {
		log.Println("Stopping Docker watcher...")
		if err := dockerWatcher.Stop(); err != nil {
			log.Printf("Error stopping Docker watcher: %v", err)
		}
	}

	// Stop all config-based proxies with timeout
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
