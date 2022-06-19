package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.ruoyu.dev/sync/channel/filter"
	"go.ruoyu.dev/sync/time/rate"
	"go.uber.org/atomic"
)

const (
	bufferSize = 1024
)

var (
	configPath = flag.String("config", "", "Path to configuration file")
	config     = &struct {
		frequency  *atomic.Float64
		filterRate *atomic.Float64
	}{
		frequency:  atomic.NewFloat64(0.0),
		filterRate: atomic.NewFloat64(0.0),
	}

	start = time.Now()
	count = 0
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func init() {
	sigint := make(chan os.Signal)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT)
	go func() {
		<-sigint
		averageFrequency := float64(count) / time.Since(start).Seconds()
		fmt.Printf("Average frequency: %f\n", averageFrequency)
		os.Exit(0)
	}()
}

func main() {
	flag.Parse()
	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Missing configuration.")
		flag.Usage()
		os.Exit(1)
	}
	updateConfig()

	allRequests := make(chan int, bufferSize)
	go func() {
		request := 0
		for range time.Tick(time.Millisecond) {
			request++
			allRequests <- request
		}
	}()

	filter := filter.NewFilter(
		allRequests,
		func(int) bool {
			return rand.Float64() < config.filterRate.Load()
		},
	)
	rateLimiter, rateLimiterError := rate.NewLimiter(config.frequency.Load())
	if rateLimiterError != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", rateLimiterError)
		os.Exit(1)
	}

	go func() {
		for range time.Tick(time.Second) {
			updateConfig()
			rateLimiterError := rateLimiter.SetFrequency(config.frequency.Load())
			if rateLimiterError != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", rateLimiterError)
				os.Exit(1)
			}
		}
	}()

	prev := time.Now()
	for request := range filter.Out() {
		if rateLimiter.Allow() {
			now := time.Now()
			interval := now.Sub(prev)
			fmt.Printf("Request %d (%v)\n", request, interval)
			prev = now
			count++
		}
	}
}

func updateConfig() {
	data, fileError := os.ReadFile(*configPath)
	if fileError != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", fileError)
		os.Exit(1)
	}
	newConfig := &struct {
		Frequency  float64 `json:"frequency"`
		FilterRate float64 `json:"filterRate"`
	}{}
	jsonError := json.Unmarshal(data, newConfig)
	if jsonError != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", jsonError)
		os.Exit(1)
	}
	config.frequency.Store(newConfig.Frequency)
	config.filterRate.Store(newConfig.FilterRate)
	fmt.Printf("Config updated: %+v\n", newConfig)
}
