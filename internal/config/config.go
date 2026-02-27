package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr            string
	KafkaBrokers        []string
	KafkaTopicSignals   string
	KafkaTopicRisk      string
	DatabaseURL         string
	AlertThreshold      float64
	AlertCooldown       time.Duration
	SimulatorTick       time.Duration
	ConsumerGroupPrefix string
}

func Load() Config {
	brokersCSV := getEnv("KAFKA_BROKERS", "localhost:19092")
	brokerParts := strings.Split(brokersCSV, ",")
	brokers := make([]string, 0, len(brokerParts))
	for _, b := range brokerParts {
		v := strings.TrimSpace(b)
		if v != "" {
			brokers = append(brokers, v)
		}
	}
	if len(brokers) == 0 {
		brokers = []string{"localhost:19092"}
	}

	tickSeconds := getEnvInt("SIMULATOR_TICK_SECONDS", 0)
	cooldownMinutes := getEnvInt("ALERT_COOLDOWN_MINUTES", 30)

	return Config{
		HTTPAddr:            getEnv("HTTP_ADDR", ":8080"),
		KafkaBrokers:        brokers,
		KafkaTopicSignals:   getEnv("KAFKA_TOPIC_SIGNALS", "signals.raw"),
		KafkaTopicRisk:      getEnv("KAFKA_TOPIC_RISK", "risk.scored"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/supplyshock?sslmode=disable"),
		AlertThreshold:      getEnvFloat("ALERT_THRESHOLD", 72),
		AlertCooldown:       time.Duration(cooldownMinutes) * time.Minute,
		SimulatorTick:       time.Duration(tickSeconds) * time.Second,
		ConsumerGroupPrefix: getEnv("CONSUMER_GROUP_PREFIX", "supplyshock"),
	}
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
