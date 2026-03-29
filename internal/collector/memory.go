package memory

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Enable   bool     `yaml:"enable"`
	EvalRate int      `yaml:"evalRate"`
	Metrics  []string `yaml:"metrics"`
}

func readMemory(metrics map[string]bool) (map[string]string, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(strings.Replace(parts[0], ":", "", 1))
		value := strings.TrimSpace(parts[1])

		if metrics[key] {
			stats[key] = value
		}
	}

	return stats, scanner.Err()
}

func (cfg Config) Collect(memc chan map[string]string) {
	ticker := time.NewTicker(time.Duration(cfg.EvalRate) * time.Millisecond)
	defer ticker.Stop()

	metrics := make(map[string]bool)
	for _, m := range cfg.Metrics {
		metrics[m] = true
	}

	for t := range ticker.C {
		stats, err := readMemory(metrics)
		if err != nil {
			log.Printf("Failed to collect memory stats: %v", err)
			continue
		}

		stats["datetime"] = t.Format(time.RFC3339)
		memc <- stats
	}
}
