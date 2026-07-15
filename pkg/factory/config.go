/*
 * DNF Configuration Factory
 */

package factory

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/free5gc/dnf/internal/logger"
	"github.com/google/uuid"
)

const (
	DnfDefaultConfigPath         = "./NFs/dnf/dnfcfg.yaml"
	DnfDefaultNfInstanceIdEnvVar = "DNF_NF_INSTANCE_ID"
	DnfSbiDefaultIPv4            = "127.0.0.101"
	DnfSbiDefaultPort            = 8000
	DnfSbiDefaultScheme          = "https"
	DnfDefaultNrfUri             = "https://127.0.0.10:8000"
	DnfDummyUriPrefix            = "/dnf-dummy/v1"
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
	NfInstanceId    string   `yaml:"nfInstanceId,omitempty" valid:"optional,uuidv4"`
	Sbi             *Sbi     `yaml:"sbi,omitempty" valid:"required"`
	ServiceNameList []string `yaml:"serviceNameList,omitempty" valid:"required"`
	NrfUri          string   `yaml:"nrfUri,omitempty" valid:"url,required"`
	NrfCertPem      string   `yaml:"nrfCertPem,omitempty" valid:"optional"`
	GroupId         string   `yaml:"groupId,omitempty" valid:"type(string),minstringlength(1)"`
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

func (c *Config) GetNfInstanceId() string {
	c.RLock()
	defer c.RUnlock()

	var nfInstanceId string

	logger.CfgLog.Debugf("Fetching nfInstanceId from env var \"%s\"", DnfDefaultNfInstanceIdEnvVar)

	if nfInstanceId = os.Getenv(DnfDefaultNfInstanceIdEnvVar); nfInstanceId == "" {
		logger.CfgLog.Debugf("No value found for \"%s\" env, fallback on config nfInstanceId : %s", DnfDefaultNfInstanceIdEnvVar, c.Configuration.NfInstanceId)
		return c.Configuration.NfInstanceId
	}

	if err := uuid.Validate(nfInstanceId); err != nil {
		logger.CfgLog.Errorf("Env var \"%s\" is ot a valid uuid, fallback on configuration nfInstanceId : %s", DnfDefaultNfInstanceIdEnvVar, c.Configuration.NfInstanceId)
		return c.Configuration.NfInstanceId
	}

	logger.CfgLog.Debug("nfInstanceId from %s : %s", DnfDefaultNfInstanceIdEnvVar, nfInstanceId)

	return nfInstanceId
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

func (c *Config) GetVersion() string {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()

	if c.Info.Version != "" {
		return c.Info.Version
	}
	return ""
}

func (c *Config) SetLogEnable(enable bool) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Enable: enable,
			Level:  "info",
		}
	} else {
		c.Logger.Enable = enable
	}
}

func (c *Config) SetLogLevel(level string) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Level: level,
		}
	} else {
		c.Logger.Level = level
	}
}

func (c *Config) SetLogReportCaller(reportCaller bool) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		c.Logger = &Logger{
			Level:        "info",
			ReportCaller: reportCaller,
		}
	} else {
		c.Logger.ReportCaller = reportCaller
	}
}

func (c *Config) GetLogEnable() bool {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.Enable
}

func (c *Config) GetLogLevel() string {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return "info"
	}
	return c.Logger.Level
}

func (c *Config) GetLogReportCaller() bool {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if c.Logger == nil {
		logger.CfgLog.Warnf("Logger should not be nil")
		return false
	}
	return c.Logger.ReportCaller
}

func (c *Config) GetSbiBindingAddr() string {
	c.RLock()
	defer c.RUnlock()
	return c.GetSbiBindingIP() + ":" + strconv.Itoa(c.GetSbiPort())
}

func (c *Config) GetSbiBindingIP() string {
	c.RLock()
	defer c.RUnlock()
	bindIP := "0.0.0.0"
	if c.Configuration == nil || c.Configuration.Sbi == nil {
		return bindIP
	}
	if c.Configuration.Sbi.BindingIPv4 != "" {
		if bindIP = os.Getenv(c.Configuration.Sbi.BindingIPv4); bindIP != "" {
			logger.CfgLog.Infof("Parsing ServerIPv4 [%s] from ENV Variable", bindIP)
		} else {
			bindIP = c.Configuration.Sbi.BindingIPv4
		}
	}
	return bindIP
}

func (c *Config) GetSbiPort() int {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Port != 0 {
		return c.Configuration.Sbi.Port
	}
	return DnfSbiDefaultPort
}

func (c *Config) GetSbiScheme() string {
	c.RLock()
	defer c.RUnlock()
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Scheme != "" {
		return c.Configuration.Sbi.Scheme
	}
	return DnfSbiDefaultScheme
}
