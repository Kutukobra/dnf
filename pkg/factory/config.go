package factory

import (
	"fmt"
	"sync"

	"github.com/asaskevich/govalidator"
)

const (
	DnfDefaultConfigPath = "./config/dnfcfg.yaml"
	DnfSbiDefaultIPv4    = "127.0.0.101"
	DnfSbiDefaultPort    = 8000
	DnfSbiDefaultScheme  = "https"
	DnfDefaultNrfUri     = "https://127.0.0.10:8000"
)

type Config struct {
	Info          *Info          `yaml:"info" valid:"required"`
	Configuration *Configuration `yaml:"configuration" valid:"required"`
	Logger        *Logger        `yaml:"logger" valid:"required"`
	sync.RWMutex
}

func (c *Config) Validate() (bool, error) {
	govalidator.TagMap["scheme"] = func(str string) bool {
		return str == "https" || str == "http"
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

func appendInvalid(err error) error {
	var errs govalidator.Errors

	if err == nil {
		return nil
	}

	es := err.(govalidator.Errors).Errors()
	for _, e := range es {
		errs = append(errs, fmt.Errorf("invalid %w", e))
	}
}
