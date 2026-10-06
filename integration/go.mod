module github.com/larsartmann/go-appkit/integration

go 1.27.1

require (
	github.com/larsartmann/cqrs-htmx/v4 v4.13.1
	github.com/larsartmann/go-appkit v0.7.0
	github.com/larsartmann/go-appkit/cqrs v0.7.0
	github.com/larsartmann/go-appkit/errorpages v0.1.1
	github.com/larsartmann/go-appkit/flightrecorderhealth v0.1.6
	github.com/larsartmann/go-appkit/health v0.1.5
	github.com/larsartmann/go-appkit/otel v0.2.0
	github.com/larsartmann/go-appkit/realtime v0.1.3
	github.com/larsartmann/go-appkit/security v0.2.0
	github.com/larsartmann/go-cqrs-lite/command/v4 v4.13.1
	github.com/larsartmann/go-cqrs-lite/decider/v4 v4.7.2
	github.com/larsartmann/go-cqrs-lite/event/v4 v4.13.1
	github.com/larsartmann/go-cqrs-lite/id/v4 v4.7.1
	github.com/larsartmann/go-cqrs-lite/metaengine/v4 v4.16.1
	github.com/larsartmann/go-cqrs-lite/query/v4 v4.10.1
	github.com/larsartmann/go-cqrs-lite/storage/memory/v4 v4.6.1
	github.com/larsartmann/go-cqrs-lite/system/v4 v4.10.2
	github.com/larsartmann/go-error-family v0.11.0
	github.com/larsartmann/go-flightrecorder v0.2.1
	github.com/larsartmann/go-health v0.5.0
	github.com/larsartmann/go-sse v0.6.2
	github.com/larsartmann/go-sse/ssetest v0.4.0
	github.com/larsartmann/httputil v1.4.1
	github.com/oklog/ulid/v2 v2.1.2
	github.com/samber/do/v2 v2.1.0
	go.opentelemetry.io/otel v1.47.0
	go.opentelemetry.io/otel/sdk v1.47.0
	go.opentelemetry.io/otel/sdk/metric v1.47.0
)

require (
	github.com/Oudwins/tailwind-merge-go v0.2.3 // indirect
	github.com/ThreeDotsLabs/watermill v1.5.3 // indirect
	github.com/a-h/templ v0.3.1020 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bits-and-blooms/bitset v1.26.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/log v1.0.0 // indirect
	github.com/charmbracelet/x/ansi v0.11.8 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/failsafe-go/failsafe-go v0.9.8 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.6 // indirect
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.31.0 // indirect
	github.com/justinas/nosurf v1.2.0 // indirect
	github.com/knadh/koanf/maps v0.1.3 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.1 // indirect
	github.com/knadh/koanf/providers/env v1.1.0 // indirect
	github.com/knadh/koanf/providers/file v1.2.1 // indirect
	github.com/knadh/koanf/v2 v2.3.7 // indirect
	github.com/larsartmann/go-branded-id v0.7.0 // indirect
	github.com/larsartmann/go-codec v0.3.1 // indirect
	github.com/larsartmann/go-cqrs-lite/claiming/v4 v4.0.2 // indirect
	github.com/larsartmann/go-cqrs-lite/commandlifecycle/projections/v4 v4.2.2 // indirect
	github.com/larsartmann/go-cqrs-lite/commandlifecycle/v4 v4.2.2 // indirect
	github.com/larsartmann/go-cqrs-lite/dedup/v4 v4.2.4 // indirect
	github.com/larsartmann/go-cqrs-lite/dispatcher/v4 v4.5.2 // indirect
	github.com/larsartmann/go-cqrs-lite/metadata/v4 v4.7.3 // indirect
	github.com/larsartmann/go-cqrs-lite/metaengine/projectionadapter/v4 v4.5.2 // indirect
	github.com/larsartmann/go-cqrs-lite/metaengine/sqliteengine/v4 v4.5.1 // indirect
	github.com/larsartmann/go-cqrs-lite/middleware/v4 v4.7.2 // indirect
	github.com/larsartmann/go-cqrs-lite/otel/v4 v4.5.2 // indirect
	github.com/larsartmann/go-cqrs-lite/projection/v4 v4.4.2 // indirect
	github.com/larsartmann/go-cqrs-lite/projectionhost/v4 v4.5.3 // indirect
	github.com/larsartmann/go-cqrs-lite/record/v4 v4.6.2 // indirect
	github.com/larsartmann/go-cqrs-lite/snapshot/v4 v4.6.1 // indirect
	github.com/larsartmann/go-cqrs-lite/watermill/v4 v4.6.4 // indirect
	github.com/larsartmann/go-datastar v0.6.2 // indirect
	github.com/larsartmann/go-datastar/static v0.6.1 // indirect
	github.com/larsartmann/go-etag/entitytag v0.6.1 // indirect
	github.com/larsartmann/go-etag/server v0.6.1 // indirect
	github.com/larsartmann/go-health-dashboard v0.10.2 // indirect
	github.com/larsartmann/go-idempotency v0.3.1 // indirect
	github.com/larsartmann/go-retry v0.7.1 // indirect
	github.com/larsartmann/go-sse/sseparse v0.2.1 // indirect
	github.com/larsartmann/templ-components v1.20.1 // indirect
	github.com/larsartmann/templ-components/datastar v1.20.1 // indirect
	github.com/larsartmann/templ-components/errorpage v1.20.1 // indirect
	github.com/larsartmann/templ-components/htmx v1.20.1 // indirect
	github.com/larsartmann/templ-components/icons v1.20.1 // indirect
	github.com/larsartmann/templ-components/utils v1.20.1 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-runewidth v0.0.30 // indirect
	github.com/maypok86/otter/v2 v2.3.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/samber/go-type-to-string v1.8.0 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xo/terminfo v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.1 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20261005173118-76772065c9b0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261005182115-fad411399dd8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261005182115-fad411399dd8 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	modernc.org/libc v1.77.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.60.1 // indirect
)

exclude (
	github.com/larsartmann/go-etag v0.1.1
	github.com/larsartmann/go-etag v0.3.1
	github.com/larsartmann/go-etag v0.4.0
	github.com/larsartmann/go-etag v0.5.0
)
