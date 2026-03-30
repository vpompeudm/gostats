package main

import (
	"fmt"
	"log"
	"os"
	"time"

	memory "gostats/internal/collector"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Memory struct {
		Collector `yaml:",inline"`
	} `yaml:"memory"`
}

type Collector struct {
	Enable        bool     `yaml:"enable"`
	EvalRate      int      `yaml:"evalRate"`
	BatchSize     int      `yaml:"batchSize"`
	FlushInterval int      `yaml:"flushInterval"`
	Metrics       []string `yaml:"metrics"`
}

type Reader func(metrics map[string]bool) (map[string]string, error)

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (collector Collector) BatchSender(chn chan map[string]string) {
	var batch []map[string]string
	// Flush every 30 seconds even if not full
	timer := time.NewTicker(time.Duration(collector.FlushInterval) * time.Millisecond)

	for {
		select {
		case stats := <-chn:
			batch = append(batch, stats)

			if len(batch) >= collector.BatchSize {
				// sendToAPI(batch)
				fmt.Printf("%v", batch)
				batch = nil
				timer.Reset(time.Duration(collector.FlushInterval) * time.Millisecond)
			}

		case <-timer.C:
			if len(batch) > 0 {
				fmt.Printf("%v", batch)
				// sendToAPI(batch)
				batch = nil
			}
		}
	}
}

func (collector Collector) Collect(chn chan map[string]string, reader Reader) {
	ticker := time.NewTicker(time.Duration(collector.EvalRate) * time.Millisecond)
	defer ticker.Stop()

	metrics := make(map[string]bool)
	for _, m := range collector.Metrics {
		metrics[m] = true
	}

	for t := range ticker.C {
		stats, err := reader(metrics)
		if err != nil {
			log.Printf("Failed to collect stats: %v", err)
			continue
		}

		stats["datetime"] = t.Format(time.RFC3339)
		chn <- stats
	}
}

func main() {
	cfg, err := LoadConfig("./configs/agent.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	memc := make(chan map[string]string, cfg.Memory.BatchSize*2)
	if cfg.Memory.Enable {
		go cfg.Memory.Collect(memc, memory.ReadMemory)
	}
	cfg.Memory.BatchSender(memc)
}
