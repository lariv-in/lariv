package p_seer_opensky

import (
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
)

// OpenSkyConfig is loaded from [lago.LagoConfig.Plugins] under key "p_seer_opensky".
//
// Example:
//
//	[Plugins.p_seer_opensky]
//	clientId = "…"
//	clientSecret = "…"
//	pollInterval = "30s"
//	# only show map markers with last_contact in this window; "0s" = no limit
//	mapLastContactMaxAge = "20s"
type OpenSkyConfig struct {
	ClientID     string `toml:"clientId"`
	ClientSecret string `toml:"clientSecret"`
	// PollInterval is a duration string, e.g. "30s", "1m". If missing or non-positive after parse, no background poller.
	PollInterval string `toml:"pollInterval"`
	// MapLastContactMaxAge limits the live map to state rows whose latest last_contact
	// is at most this far behind now (e.g. "20s", "5m"). Empty => "1m". "0" or "0s" => no time filter.
	MapLastContactMaxAge string `toml:"mapLastContactMaxAge"`

	pollEvery                  time.Duration
	mapLastContactMaxAgeParsed time.Duration
}

// PollEvery returns the parsed poll interval, or 0 if inactive.
func (c *OpenSkyConfig) PollEvery() time.Duration {
	if c == nil {
		return 0
	}
	return c.pollEvery
}

// MapLastContactWindow returns the parsed [MapLastContactMaxAge] string.
// Returns 0 to apply no time ceiling (HAVING last_contact >= 0).
func (c *OpenSkyConfig) MapLastContactWindow() time.Duration {
	if c == nil {
		return 0
	}
	return c.mapLastContactMaxAgeParsed
}

func (c *OpenSkyConfig) PostConfig() {
	if c == nil {
		return
	}
	if c.PollInterval == "" {
		c.PollInterval = "30s"
	}
	d, err := time.ParseDuration(c.PollInterval)
	if err != nil {
		c.pollEvery = 0
		return
	}
	c.pollEvery = d

	age := c.MapLastContactMaxAge
	if age == "" {
		age = "1m"
	}
	ma, err := time.ParseDuration(age)
	if err != nil {
		slog.Warn("p_seer_opensky: mapLastContactMaxAge invalid, using no time limit on map", "value", c.MapLastContactMaxAge, "error", err)
		c.mapLastContactMaxAgeParsed = 0
		return
	}
	c.mapLastContactMaxAgeParsed = ma
}

var Config = &OpenSkyConfig{}

func init() {
	lago.RegistryConfig.Register("p_seer_opensky", Config)
}
