package main

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"

	_ "github.com/lariv-in/lago/plugins/p_dashboard"
	_ "github.com/lariv-in/lago/plugins/p_filesystem"
	_ "github.com/lariv-in/lago/plugins/p_google_genai"
	_ "github.com/lariv-in/lago/plugins/p_pwa"
	_ "github.com/lariv-in/lago/plugins/p_seer_assistant"
	_ "github.com/lariv-in/lago/plugins/p_seer_deepsearch"
	_ "github.com/lariv-in/lago/plugins/p_seer_gdelt"
	_ "github.com/lariv-in/lago/plugins/p_seer_intel"
	_ "github.com/lariv-in/lago/plugins/p_seer_opensky"
	_ "github.com/lariv-in/lago/plugins/p_seer_aisstream"
	_ "github.com/lariv-in/lago/plugins/p_seer_reddit"
	_ "github.com/lariv-in/lago/plugins/p_seer_runners"
	_ "github.com/lariv-in/lago/plugins/p_seer_websites"
	_ "github.com/lariv-in/lago/plugins/p_users"
)

func main() {
	config, err := lago.LoadConfigFromFile("seer.toml")
	if err != nil {
		panic(err)
	}
	if err := lago.Start(config); err != nil {
		slog.Error(err.Error())
	}
}
