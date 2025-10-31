package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	RedisAddr     string
	RedisPass     string
	Port          string
	DataGovAPIKey string
	CronSchedule  string
	FYList        []string
	HTTPTimeout   time.Duration
}

func Load() *Config {
	c := &Config{
		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:     getEnv("REDIS_PASS", ""),
		Port:          getEnv("PORT", "8080"),
		DataGovAPIKey: getEnv("DATA_GOV_API_KEY", "579b464db66ec23bdd000001cdd3946e44ce4aad7209ff7b23ac571b"), //default api key with 10 pagination
		CronSchedule:  getEnv("CRON_SCHEDULE", "0 2 * * *"),                                                   // daily 02:00 default
		HTTPTimeout:   50 * time.Second,
	}

	if v := os.Getenv("FY_LIST"); v != "" {
		c.FYList = strings.Split(v, ",")
	} else {
		// generate rolling last 6 FY start years (e.g. "2020-2021")
		cy := time.Now().Year()
		fy := make([]string, 0, 6)
		for i := 0; i < 6; i++ {
			start := cy - i
			end := start + 1
			fy = append(fy, fmt.Sprintf("%d-%d", start, end))
		}
		c.FYList = fy // newest first
	}
	return c
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
