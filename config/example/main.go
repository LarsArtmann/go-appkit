// Command example demonstrates appkit/config: env-first struct tags with
// defaults, a file layer, and classified field-path errors.
//
// Try it:
//
//	GOWORK=off GOTOOLCHAIN=go1.27.2 go run .                       # missing required → Rejection
//	APP_API_KEY=secret GOWORK=off GOTOOLCHAIN=go1.27.2 go run .   # defaults
//	APP_API_KEY=s APP_TTL=5m APP_ADDR=:9000 go run .              # env overrides
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/larsartmann/go-appkit/config"
)

// Config is the resolved application configuration.
type Config struct {
	Addr     string        `env:"APP_ADDR" default:":8080"`
	LogLevel string        `env:"APP_LOG_LEVEL" default:"info"`
	APIKey   string        `env:"APP_API_KEY,required"`
	TTL      time.Duration `env:"APP_TTL" default:"60s"`
	Hosts    []string      `env:"APP_HOSTS"`
}

func main() {
	cfg, err := config.Load[Config]()
	if err != nil {
		log.Fatal(err)
	}

	keys, _ := config.Keys[Config]()

	fmt.Printf("loaded: %+v\n", cfg)
	fmt.Printf("declared env keys: %v\n", keys)
}
