package config

import (
	"os"
	"strings"
)

// LoadDotEnv loads environment variables from a key=value file if not already set in the process environment.
func LoadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "\"'")
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

// UpdateDotEnv sets or updates an environment variable in the process and persists it to filepath.
func UpdateDotEnv(filepath, key, val string) error {
	os.Setenv(key, val)
	data, err := os.ReadFile(filepath)
	var lines []string
	found := false
	if err == nil {
		rawLines := strings.Split(string(data), "\n")
		for _, line := range rawLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, key+"=") {
				lines = append(lines, key+"=\""+val+"\"")
				found = true
			} else {
				lines = append(lines, line)
			}
		}
	}
	if !found {
		lines = append(lines, key+"=\""+val+"\"")
	}
	return os.WriteFile(filepath, []byte(strings.Join(lines, "\n")), 0644)
}

// IsConfiguredKey returns true if the key is non-empty and not a known dummy/placeholder template string.
func IsConfiguredKey(val string) bool {
	val = strings.TrimSpace(val)
	if val == "" {
		return false
	}
	lower := strings.ToLower(val)
	if strings.Contains(lower, "replace") ||
		strings.Contains(lower, "your_actual") ||
		strings.Contains(lower, "todo") ||
		strings.Contains(lower, "dummy") ||
		strings.Contains(lower, "example") ||
		strings.Contains(lower, "test_key") {
		return false
	}
	return true
}
