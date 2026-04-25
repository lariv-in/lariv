module github.com/lariv-in/lago/deployments/seer

go 1.26.1

require (
	github.com/lariv-in/lago/lago v0.5.10
	github.com/lariv-in/lago/plugins/p_dashboard v0.5.10
	github.com/lariv-in/lago/plugins/p_filesystem v0.5.10
	github.com/lariv-in/lago/plugins/p_google_genai v0.0.0-00010101000000-000000000000
	github.com/lariv-in/lago/plugins/p_livereloading v0.0.0-20260425015217-4dcf321277c2
	github.com/lariv-in/lago/plugins/p_pwa v0.5.10
	github.com/lariv-in/lago/plugins/p_seer_assistant v0.0.0-00010101000000-000000000000
	github.com/lariv-in/lago/plugins/p_seer_deepsearch v0.0.0-20260419161526-9d250d818084
	github.com/lariv-in/lago/plugins/p_seer_intel v0.0.0-20260421170520-3ca65872f4a4
	github.com/lariv-in/lago/plugins/p_seer_opensky v0.0.0
	github.com/lariv-in/lago/plugins/p_seer_reddit v0.0.0-20260421170520-3ca65872f4a4
	github.com/lariv-in/lago/plugins/p_seer_runners v0.0.0-00010101000000-000000000000
	github.com/lariv-in/lago/plugins/p_seer_websites v0.0.0-00010101000000-000000000000
	github.com/lariv-in/lago/plugins/p_users v0.5.10
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/bigquery v1.76.0 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.7.0 // indirect
	codeberg.org/readeck/go-readability/v2 v2.1.1 // indirect
	github.com/JohannesKaufmann/dom v0.2.0 // indirect
	github.com/JohannesKaufmann/html-to-markdown/v2 v2.5.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-rod/rod v0.116.2 // indirect
	github.com/go-shiori/dom v0.0.0-20230515143342-73569d674e1c // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.14 // indirect
	github.com/googleapis/gax-go/v2 v2.21.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/lariv-in/lago/syncmap v0.0.0-20260421131401-4ff0a5e51b63 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/ysmood/fetchup v0.2.3 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.42.3 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.9.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.67.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/telemetry v0.0.0-20260311193753-579e4da9a98c // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.276.0 // indirect
	google.golang.org/genai v1.54.0 // indirect
	google.golang.org/genproto v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260401024825-9d38bb4040a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
	google.golang.org/grpc v1.80.0 // indirect
)

replace github.com/lariv-in/lago/plugins/p_seer_assistant => ../../plugins/p_seer_assistant

replace github.com/lariv-in/lago/plugins/p_google_genai => ../../plugins/p_google_genai

replace github.com/lariv-in/lago/plugins/p_seer_intel => ../../plugins/p_seer_intel

replace github.com/lariv-in/lago/plugins/p_seer_reddit => ../../plugins/p_seer_reddit

replace github.com/lariv-in/lago/plugins/p_seer_runners => ../../plugins/p_seer_runners

replace github.com/lariv-in/lago/plugins/p_seer_websites => ../../plugins/p_seer_websites

require (
	charm.land/bubbles/v2 v2.1.0 // indirect
	charm.land/bubbletea/v2 v2.0.2 // indirect
	charm.land/huh/v2 v2.0.3 // indirect
	charm.land/lipgloss/v2 v2.0.2 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/catppuccin/go v0.3.0 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20260330092749-0f94982c930b // indirect
	github.com/charmbracelet/x/ansi v0.11.6 // indirect
	github.com/charmbracelet/x/exp/ordered v0.1.0 // indirect
	github.com/charmbracelet/x/exp/strings v0.1.0 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/gomarkdown/markdown v0.0.0-20260217112301-37c66b85d6ab // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lariv-in/lago/components v0.5.10 // indirect
	github.com/lariv-in/lago/getters v0.5.10 // indirect
	github.com/lariv-in/lago/plugins/p_seer_gdelt v0.0.0-00010101000000-000000000000
	github.com/lariv-in/lago/registry v0.5.10 // indirect
	github.com/lariv-in/lago/views v0.5.10 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.0 // indirect
	github.com/mattn/go-runewidth v0.0.22 // indirect
	github.com/mattn/go-sqlite3 v1.14.40 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/nyaruka/phonenumbers v1.7.1 // indirect
	github.com/pgvector/pgvector-go v0.3.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gorm.io/datatypes v1.2.7 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/driver/sqlite v1.6.0 // indirect
	gorm.io/gorm v1.31.1 // indirect
	maragu.dev/gomponents v1.3.0 // indirect
)

replace github.com/lariv-in/lago/plugins/p_seer_opensky => ../../plugins/p_seer_opensky

replace github.com/lariv-in/lago/plugins/p_seer_aisstream => ../../plugins/p_seer_aisstream

replace github.com/lariv-in/lago/plugins/p_seer_gdelt => ../../plugins/p_seer_gdelt
