package p_seer_aisstream

import (
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
)

type AISStreamConfig struct {
	APIKey               string `toml:"apiKey"`
	StreamURL            string `toml:"streamURL"`
	Enabled              bool   `toml:"enabled"`
	MapLastContactMaxAge string `toml:"mapLastContactMaxAge"`
	MapRefreshInterval   string `toml:"mapRefreshInterval"`

	mapLastContactMaxAgeParsed time.Duration
	mapRefreshIntervalParsed   time.Duration
}

func (c *AISStreamConfig) PostConfig() {
	if c == nil {
		return
	}
	if c.StreamURL == "" {
		c.StreamURL = "wss://stream.aisstream.io/v0/stream"
	}
	if c.MapLastContactMaxAge == "" {
		c.MapLastContactMaxAge = "10m"
	}
	age, err := time.ParseDuration(c.MapLastContactMaxAge)
	if err != nil {
		slog.Warn("p_seer_aisstream: invalid mapLastContactMaxAge, using no time limit", "value", c.MapLastContactMaxAge, "error", err)
		age = 0
	}
	c.mapLastContactMaxAgeParsed = age

	if c.MapRefreshInterval == "" {
		c.MapRefreshInterval = "5s"
	}
	refresh, err := time.ParseDuration(c.MapRefreshInterval)
	if err != nil {
		slog.Warn("p_seer_aisstream: invalid mapRefreshInterval, disabling map polling", "value", c.MapRefreshInterval, "error", err)
		refresh = 0
	}
	c.mapRefreshIntervalParsed = refresh
}

func (c *AISStreamConfig) MapLastContactWindow() time.Duration {
	if c == nil {
		return 0
	}
	return c.mapLastContactMaxAgeParsed
}

func (c *AISStreamConfig) MapRefreshEvery() time.Duration {
	if c == nil {
		return 0
	}
	return c.mapRefreshIntervalParsed
}

var Config = &AISStreamConfig{Enabled: true}

func init() {
	lago.RegistryConfig.Register("p_seer_aisstream", Config)
}
