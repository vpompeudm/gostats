package memory

import (
	"bufio"
	"os"
	"strings"
)

func ReadMemory(metrics map[string]bool) (map[string]string, error) {
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
