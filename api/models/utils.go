package models

import (
	"encoding/json"
	"os"
	"strconv"
	"time"
)

func ConvertType(input, output interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, output)
	if err != nil {
		return err
	}
	return nil
}

func CacheDuration() time.Duration {
	cacheDuration := os.Getenv("CACHE_DURATION")
	if cacheDuration == "" {
		cacheDuration = "1"
	}
	duration, _ := strconv.ParseUint(cacheDuration, 10, 64)
	return time.Minute * time.Duration(duration)
}
