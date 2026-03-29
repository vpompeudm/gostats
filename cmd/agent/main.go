package main

import (
	"fmt"
	"log"
	"os"

	memory "gostats/internal/collector"

	"go.yaml.in/yaml/v4"
)

type Collectors struct {
	Memory struct {
		memory.Config `yaml:",inline"`
	} `yaml:"memory"`
}

type Metric interface {
	Collect() map[string]float64
}

func LoadCollectors(path string) (*Collectors, error) {
	cfg := &Collectors{}

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

func main() {
	cfg, err := LoadCollectors("./configs/agent.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	if cfg.Memory.Enable {
		memc := make(chan map[string]string)
		go cfg.Memory.Collect(memc)
		for range 10 {
			fmt.Println(<-memc)
		}
	}
}
