package config

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	values map[string]any
}

var (
	manager *Manager
	once    sync.Once
)

func LoadFromBytes(bytes []byte) error {
	values := map[string]any{}
	if err := yaml.Unmarshal(bytes, &values); err != nil {
		return err
	}
	manager = &Manager{values: values}
	return nil
}

func LoadFromPath(path string) error {
	var loadErr error
	once.Do(func() {
		bytes, err := os.ReadFile(path)
		if err != nil {
			bytes, err = exec.Command("cat", path).Output()
			if err != nil {
				loadErr = err
				return
			}
		}
		loadErr = LoadFromBytes(bytes)
	})

	return loadErr
}

func MustLoadFromEnv() error {
	if inline := os.Getenv("CONFIG_INLINE"); inline != "" {
		var loadErr error
		once.Do(func() {
			loadErr = LoadFromBytes([]byte(inline))
		})
		return loadErr
	}

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return errors.New("CONFIG_PATH or CONFIG_INLINE is required")
	}

	return LoadFromPath(path)
}

func GetString(path string) string {
	value := Get(path)
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(toString(typed)), "\n", "")), "]"), "["), "\"", ""))
	}
}

func GetBool(path string) bool {
	value := Get(path)
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(typed)
		return parsed
	default:
		return false
	}
}

func GetInt(path string) int {
	value := Get(path)
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	default:
		return 0
	}
}

func GetInt64(path string) int64 {
	value := Get(path)
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case string:
		parsed, _ := strconv.ParseInt(typed, 10, 64)
		return parsed
	default:
		return 0
	}
}

func GetStringSlice(path string) []string {
	value := Get(path)
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, toString(item))
		}
		return result
	default:
		return nil
	}
}

func Get(path string) any {
	if manager == nil {
		return nil
	}

	parts := strings.Split(path, ".")
	current := any(manager.values)
	for _, part := range parts {
		mapValue, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		current, ok = mapValue[part]
		if !ok {
			return nil
		}
	}

	return current
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return ""
	}
}
