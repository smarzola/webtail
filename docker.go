package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	labelEnabled            = "webtail.enabled"
	labelProtocol           = "webtail.protocol"
	labelPort               = "webtail.port"
	labelNodeName           = "webtail.node_name"
	labelPassHostHeader     = "webtail.pass_host_header"
	labelTrustForwardHeader = "webtail.trust_forward_header"

	defaultProtocol = "http"
)

// DockerWatcher watches for Docker container events and manages proxies
type DockerWatcher struct {
	client        *client.Client
	tsConfig      *TailscaleConfig
	dockerNetwork string
	proxies       map[string]*Proxy // containerID -> Proxy
	mu            sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewDockerWatcher creates a new Docker event watcher
func NewDockerWatcher(tsConfig *TailscaleConfig, dockerConfig *DockerConfig) (*DockerWatcher, error) {
	// Build client options - start with config values, then let environment variables override
	var opts []client.Opt

	// Apply config values first (lowest priority)
	if dockerConfig.Host != "" {
		opts = append(opts, client.WithHost(dockerConfig.Host))
	}
	if dockerConfig.APIVersion != "" {
		opts = append(opts, client.WithVersion(dockerConfig.APIVersion))
	}
	if dockerConfig.CertPath != "" {
		tlsVerify := boolValue(dockerConfig.TLSVerify, false)
		if tlsVerify {
			opts = append(opts, client.WithTLSClientConfig(
				dockerConfig.CertPath+"/ca.pem",
				dockerConfig.CertPath+"/cert.pem",
				dockerConfig.CertPath+"/key.pem",
			))
		}
	}

	// Apply environment variables last (highest priority - overrides config)
	opts = append(opts, client.FromEnv, client.WithAPIVersionNegotiation())

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DockerWatcher{
		client:        cli,
		tsConfig:      tsConfig,
		dockerNetwork: dockerConfig.Network,
		proxies:       make(map[string]*Proxy),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start begins watching for Docker events
func (dw *DockerWatcher) Start() error {
	// First, scan existing containers
	if err := dw.scanExistingContainers(); err != nil {
		log.Printf("Warning: failed to scan existing containers: %v", err)
	}

	// Set up event filters for container start events
	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")
	filterArgs.Add("event", "start")

	eventsChan, errChan := dw.client.Events(dw.ctx, events.ListOptions{
		Filters: filterArgs,
	})

	dw.wg.Add(1)
	go func() {
		defer dw.wg.Done()
		log.Println("Docker watcher started, listening for container events...")

		for {
			select {
			case <-dw.ctx.Done():
				log.Println("Docker watcher stopping...")
				return
			case err := <-errChan:
				if err != nil && dw.ctx.Err() == nil {
					log.Printf("Docker events error: %v", err)
				}
				return
			case event := <-eventsChan:
				dw.handleEvent(event)
			}
		}
	}()

	return nil
}

// scanExistingContainers checks running containers for webtail labels
func (dw *DockerWatcher) scanExistingContainers() error {
	containers, err := dw.client.ContainerList(dw.ctx, container.ListOptions{
		All: false, // Only running containers
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		if err := dw.handleContainer(c.ID); err != nil {
			log.Printf("Error handling existing container %s: %v", c.ID[:12], err)
		}
	}

	return nil
}

// handleEvent processes a Docker event
func (dw *DockerWatcher) handleEvent(event events.Message) {
	if event.Type != events.ContainerEventType {
		return
	}

	switch event.Action {
	case "start":
		if err := dw.handleContainer(event.Actor.ID); err != nil {
			log.Printf("Error handling container %s: %v", event.Actor.ID[:12], err)
		}
	}
}

// handleContainer inspects a container and starts a proxy if enabled
func (dw *DockerWatcher) handleContainer(containerID string) error {
	// Inspect the container to get full labels and container name
	inspect, err := dw.client.ContainerInspect(dw.ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	labels := inspect.Config.Labels

	// Check if webtail is enabled
	enabledStr, hasEnabled := labels[labelEnabled]
	if !hasEnabled || strings.ToLower(enabledStr) != "true" {
		return nil // Not enabled, skip
	}

	// Get container name (remove leading slash)
	containerName := strings.TrimPrefix(inspect.Name, "/")

	// Get port from label or detect from exposed ports
	port := labels[labelPort]
	if port == "" {
		// Auto-detect port from container's exposed ports (use lowest)
		detectedPort := getLowestExposedPort(inspect.Config.ExposedPorts)
		if detectedPort == "" {
			log.Printf("Container %s has webtail.enabled=true but no webtail.port label and no exposed ports", containerID[:12])
			return nil
		}
		port = detectedPort
		log.Printf("Container %s: auto-detected port %s (lowest exposed port)", containerID[:12], port)
	}

	// Get node name from label or default to container name
	nodeName := labels[labelNodeName]
	if nodeName == "" {
		nodeName = containerName
		log.Printf("Container %s: using container name %q as node name", containerID[:12], nodeName)
	}

	// Get optional labels with defaults
	protocol := labels[labelProtocol]
	if protocol == "" {
		protocol = defaultProtocol
	}
	passHostHeader := parseBoolLabel(labels[labelPassHostHeader], false)
	trustForwardHeader := parseBoolLabel(labels[labelTrustForwardHeader], false)

	// Build target URL dynamically: {protocol}://{container_name}.{docker_network}:{port}
	target := fmt.Sprintf("%s://%s.%s:%s", protocol, containerName, dw.dockerNetwork, port)

	// Check if we already have a proxy for this container
	dw.mu.Lock()
	if _, exists := dw.proxies[containerID]; exists {
		dw.mu.Unlock()
		log.Printf("Proxy already exists for container %s (%s)", containerID[:12], nodeName)
		return nil
	}
	dw.mu.Unlock()

	// Create service config from labels
	serviceConfig := &ServiceConfig{
		Target:             target,
		NodeName:           nodeName,
		PassHostHeader:     &passHostHeader,
		TrustForwardHeader: &trustForwardHeader,
	}

	log.Printf("Container %s started with webtail enabled: %s -> %s",
		containerID[:12], nodeName, target)

	// Create and start proxy
	proxy := NewProxy(serviceConfig, dw.tsConfig)

	dw.wg.Add(1)
	go func() {
		defer dw.wg.Done()

		if err := proxy.Start(); err != nil {
			log.Printf("Failed to start proxy for container %s (%s): %v",
				containerID[:12], nodeName, err)
			return
		}

		dw.mu.Lock()
		dw.proxies[containerID] = proxy
		dw.mu.Unlock()

		log.Printf("Started proxy for container %s (%s)", containerID[:12], nodeName)

		// Watch for container stop/die events
		dw.watchContainerStop(containerID, nodeName)
	}()

	return nil
}

// watchContainerStop monitors for when a container stops
func (dw *DockerWatcher) watchContainerStop(containerID, nodeName string) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")
	filterArgs.Add("container", containerID)
	filterArgs.Add("event", "stop")
	filterArgs.Add("event", "die")
	filterArgs.Add("event", "kill")

	eventsChan, errChan := dw.client.Events(dw.ctx, events.ListOptions{
		Filters: filterArgs,
	})

	for {
		select {
		case <-dw.ctx.Done():
			return
		case err := <-errChan:
			if err != nil && dw.ctx.Err() == nil {
				log.Printf("Error watching container %s: %v", containerID[:12], err)
			}
			return
		case event := <-eventsChan:
			if event.Action == "stop" || event.Action == "die" || event.Action == "kill" {
				log.Printf("Container %s (%s) stopped, shutting down proxy",
					containerID[:12], nodeName)
				dw.stopProxy(containerID)
				return
			}
		}
	}
}

// stopProxy stops and removes a proxy for a container
func (dw *DockerWatcher) stopProxy(containerID string) {
	dw.mu.Lock()
	proxy, exists := dw.proxies[containerID]
	if exists {
		delete(dw.proxies, containerID)
	}
	dw.mu.Unlock()

	if exists && proxy != nil {
		if err := proxy.Stop(); err != nil {
			log.Printf("Error stopping proxy for container %s: %v", containerID[:12], err)
		}
	}
}

// Stop gracefully shuts down the Docker watcher and all managed proxies
func (dw *DockerWatcher) Stop() error {
	dw.cancel()

	// Stop all managed proxies
	dw.mu.Lock()
	proxiesToStop := make([]*Proxy, 0, len(dw.proxies))
	for _, proxy := range dw.proxies {
		proxiesToStop = append(proxiesToStop, proxy)
	}
	dw.proxies = make(map[string]*Proxy)
	dw.mu.Unlock()

	var stopWg sync.WaitGroup
	for _, proxy := range proxiesToStop {
		stopWg.Add(1)
		go func(p *Proxy) {
			defer stopWg.Done()
			if err := p.Stop(); err != nil {
				log.Printf("Error stopping proxy: %v", err)
			}
		}(proxy)
	}
	stopWg.Wait()

	// Close Docker client
	if dw.client != nil {
		dw.client.Close()
	}

	dw.wg.Wait()
	return nil
}

// GetProxies returns the current list of managed proxies
func (dw *DockerWatcher) GetProxies() []*Proxy {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	proxies := make([]*Proxy, 0, len(dw.proxies))
	for _, p := range dw.proxies {
		proxies = append(proxies, p)
	}
	return proxies
}

// parseBoolLabel parses a string label as boolean with a default value
func parseBoolLabel(value string, defaultVal bool) bool {
	if value == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}
	return b
}

// getLowestExposedPort returns the lowest port number from the container's exposed ports
func getLowestExposedPort(exposedPorts nat.PortSet) string {
	if len(exposedPorts) == 0 {
		return ""
	}

	// Collect all port numbers
	var ports []int
	for port := range exposedPorts {
		portNum := port.Int()
		if portNum > 0 {
			ports = append(ports, portNum)
		}
	}

	if len(ports) == 0 {
		return ""
	}

	// Sort and return the lowest
	sort.Ints(ports)
	return strconv.Itoa(ports[0])
}

