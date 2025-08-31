package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/vulcand/oxy/forward"
	"tailscale.com/tsnet"
)

// Proxy represents a single service proxy
type Proxy struct {
	config    *ServiceConfig
	tsConfig  *TailscaleConfig
	server    *tsnet.Server
	forwarder http.Handler
	listener  net.Listener
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewProxy creates a new proxy instance for a service
func NewProxy(serviceConfig *ServiceConfig, tsConfig *TailscaleConfig) *Proxy {
	_, cancel := context.WithCancel(context.Background())

	return &Proxy{
		config:   serviceConfig,
		tsConfig: tsConfig,
		cancel:   cancel,
	}
}

// Start initializes and starts the proxy server
func (p *Proxy) Start() error {

	basedir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}

	// Create tsnet server
	p.server = &tsnet.Server{
		Hostname:  p.config.NodeName,
		AuthKey:   p.tsConfig.AuthKey,
		Ephemeral: p.tsConfig.Ephemeral,
		Logf:      log.Printf,
		Dir:       fmt.Sprintf("%s/webtail/%s", basedir, p.config.NodeName),
	}

	// Start the tsnet server
	if err := p.server.Start(); err != nil {
		return fmt.Errorf("failed to start tsnet server for %s: %w", p.config.NodeName, err)
	}

	// Create oxy forwarder
	fwd, err := forward.New()
	if err != nil {
		p.server.Close()
		return fmt.Errorf("failed to create forwarder for %s: %w", p.config.NodeName, err)
	}
	p.forwarder = fwd

	// Create listener on the tailnet
	listener, err := p.server.ListenTLS("tcp", ":443")
	if err != nil {
		p.server.Close()
		return fmt.Errorf("failed to create listener for %s: %w", p.config.NodeName, err)
	}
	p.listener = listener

	// Create HTTP server
	server := &http.Server{
		Handler: http.HandlerFunc(p.handleRequest),
	}

	// Start serving in a goroutine
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		log.Printf("Starting proxy for %s -> %s",
			p.config.NodeName, p.config.Target)

		if err := server.Serve(p.listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error for %s: %v", p.config.NodeName, err)
		}
	}()

	return nil
}

// handleRequest forwards the request to the upstream service
func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the target URL
	targetURL, err := url.Parse(p.config.Target)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		log.Printf("Failed to parse target URL %s: %v", p.config.Target, err)
		return
	}

	// Preserve the original scheme if not specified
	if targetURL.Scheme == "" {
		targetURL.Scheme = "http"
	}

	// Update path and query from the incoming request
	targetURL.Path = r.URL.Path
	targetURL.RawQuery = r.URL.RawQuery

	// Update the request URL
	r.URL = targetURL
	r.Host = targetURL.Host

	// Forward the request
	p.forwarder.ServeHTTP(w, r)
}

// Stop gracefully shuts down the proxy
func (p *Proxy) Stop() error {
	p.cancel()

	if p.listener != nil {
		p.listener.Close()
	}

	if p.server != nil {
		p.server.Close()
	}

	// Wait for the serving goroutine to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Proxy for %s stopped", p.config.NodeName)
		return nil
	case <-time.After(10 * time.Second):
		log.Printf("Timeout waiting for proxy %s to stop", p.config.NodeName)
		return fmt.Errorf("timeout stopping proxy for %s", p.config.NodeName)
	}
}
