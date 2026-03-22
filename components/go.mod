module github.com/lariv-in/lago/components

go 1.26.1

require (
	github.com/gomarkdown/markdown v0.0.0-20260217112301-37c66b85d6ab
	github.com/lariv-in/lago/getters v0.0.0
	github.com/lariv-in/lago/registry v0.0.0
	maragu.dev/gomponents v1.2.0
)

replace (
	github.com/lariv-in/lago/getters => ../getters
	github.com/lariv-in/lago/registry => ../registry
)

require (
	github.com/nyaruka/phonenumbers v1.6.11
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
