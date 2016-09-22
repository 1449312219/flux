package instance

import (
	"github.com/weaveworks/fluxy"
)

type ServiceConfig struct {
	Automated bool `json:"automation"`
}

type Config struct {
	Services map[flux.ServiceID]ServiceConfig `json:"services"`
}

func MakeConfig() Config {
	return Config{
		Services: map[flux.ServiceID]ServiceConfig{},
	}
}

type UpdateFunc func(config Config) (Config, error)

type DB interface {
	Update(instance flux.InstanceID, update UpdateFunc) error
	Get(instance flux.InstanceID) (Config, error)
}
