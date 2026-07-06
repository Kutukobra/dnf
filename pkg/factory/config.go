package factory

import (
	"fmt"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
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

	if configuration := c.Configuration; configuration != nil {
		if result, err := configuration.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	NfInstanceId string `yaml:"nfInstanceId,omitempty" valid:"optional,uuidv4"`
	Sbi          *Sbi   `yaml:"sbi,omitempty" valid:"required"`
	NrfUri       string `yaml:"nrfUri,omitempty" valid:"url,required"`
	GroupId      string `yaml:"groupId,omitempty" valid:"type(string),minstringlength(1)"`
}

func (c *Configuration) validate() (bool, error) {
	if c.NfInstanceId == "" {
		c.NfInstanceId = uuid.New().String()
	}

	if sbi := c.Sbi; sbi != nil {
		if result, err := sbi.validate(); err != nil {
			return result, err
		}
	}

	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
}

type Logger struct {
	Enable       bool   `yaml:"enable" valid:"type(bool)"`
	Level        string `yaml:"level" valid:"required,in(trace|debug|info|warn|error|fatal|panic)"`
	ReportCaller bool   `yaml:"reportCaller" valid:"type(bool)"`
}

type Sbi struct {
	Scheme       string `yaml:"scheme" valid:"scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"host,required"` // IP that is registered at NRF.
	BindingIPv4  string `yaml:"bindingIPv4,omitempty" valid:"host,required"`  // IP used to run the server in the node.
	Port         int    `yaml:"port,omitempty" valid:"port,required"`
}

func (c *Sbi) validate() (bool, error) {
	result, err := govalidator.ValidateStruct(c)
	return result, appendInvalid(err)
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

	return error(errs)
}
