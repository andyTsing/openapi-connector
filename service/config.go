package service

import (
	"encoding/json"
	"fmt"

	"github.com/go-ocf/openapi-connector/store"

	"github.com/go-ocf/kit/net/grpc"
)

//Config represent application configuration
type Config struct {
	grpc.Config
	AuthHost              string `envconfig:"AUTH_HOST"  default:"127.0.0.1:7000"`
	ResourceAggregateHost string `envconfig:"RESOURCE_AGGREGATE_HOST"  default:"127.0.0.1:9083"`
	FQDN                  string `envconfig:"FQDN" default:"openapi.pluggedin.cloud"`
	OAuthCallback         string `envconfig:"OAUTH_CALLBACK" required:"true"`
	EventsURL             string `envconfig:"EVENTS_URL" required:"true"`
	OriginCloud           store.LinkedCloud
}

//String return string representation of Config
func (c Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return fmt.Sprintf("config: \n%v\n", string(b))
}
