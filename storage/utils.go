package storage

import (
	"encoding/json"
	"fmt"
)

func GenerateCacheKey(prefix string, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

func Serialize(value interface{}) (string, error) {
	json, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

func Deserialize(value string, v interface{}) error {
	return json.Unmarshal([]byte(value), v)
}
