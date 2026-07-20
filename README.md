# Basic NF Module Implementation

:::    spoiler Table of Contents
[TOC]
:::

## Objectives
Creating a NF function that:
- Has its own endpoint that returns 'Hello World'
- Register itself into the NRF
- Fetch subscriber data (NSSAI) from UDM

## Components of an NF Module

### SBI
Service-Based Interface is an RESTful API-driven framework that allows Control Plane NFs to interract over HTTP/2. 

#### Consumer
Interfaces that calls other NFs' SBI.

#### Processor
Interfaces that provides services for other NFs.

### NRF Registration
An NF module registers into the NRF through the NRF's SBI. It uses the NFManagement OpenAPI specification.

### Config
A config `.yaml` file read by Golang yaml library and validated by Govalidator.

### Context
Contains all the informations of an NF module stored as struct used along its lifecycle. A lot of it is taken from the config file.

## Structure of an NF Module
```
dnf/
├── cmd
│   └── main.go
├── dnfcfg.yaml
├── go.mod
├── go.sum
├── internal
│   ├── context
│   │   └── context.go
│   ├── logger
│   │   └── logger.go
│   └── sbi
│       ├── api_dummy.go
│       ├── consumer
│       │   ├── consumer.go
│       │   ├── nrf_service.go
│       │   ├── nrf_service_test.go
│       │   └── udm_service.go
│       ├── processor
│       │   ├── dummy.go
│       │   └── processor.go
│       ├── routes.go
│       └── server.go
├── pkg
│   ├── app
│   │   └── app.go
│   ├── factory
│   │   ├── config.go
│   │   └── factory.go
│   └── service
│       └── init.go
```

## Building an NF Module
To build the NF, the folder name `dnf` is added to `free5gc/makefile`.
```makefile=
NF = $(GO_NF)
GO_NF = amf ausf bsf nrf nssf pcf smf udm udr n3iwf upf chf tngf nef dnf
```

Running is simply done by command `./bin/dnf` in `free5gc` dir.


## Creating an NF Module
The NF `AUSF` is observed as it is one of the simpler control network function, giving a general idea. I made simple network module called `DNF` (Dummy Network Function) as a study.

### Entrypoint
A Free5GC NF utilizes `github.com/urfave/cli/v2` as a top `app.NewApp()` CLI layer to parse flags. The function itself starts with the `app.Action`.

#### Main
`cmd/main.go`
```go=
func main() {
    defer func() {
        if p := recover(); p != nil {
            // Print stack for panic to log. Fatalf() will let program exit.
            logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
        }
    }()

    app := cli.NewApp()
    app.Name = "dnf"
    app.Usage = "5G Dummy Network Function (DNF)"
    app.Action = action
    app.Flags = []cli.Flag{
        &cli.StringSliceFlag{
            Name: "log",
            Aliases: []string{"l"},
            Usage: "Output NF log to `FILE`",
        }
    }

    if err := app.Run(os.Args); err != nil {
        logger.MainLog.Errorf("DNF Run error: %v\n", err)
    }
}
```
The function initiates a `cli` app with the descriptions and flags to it. A deferred function is created in case of a panicking goroutine. The `cli` app is then run wiht the OS arguments with `app.Run(os.Args)`.

#### Action
`cmd/main.go`
```go=
func action(cliCtx *cli.Context) error {
    logger.MainLog.Infoln("DNF version: ", version.GetVersion())

    ctx, cancel := context.WithCancel(context.Background())
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigCh
        cancel()
    }()

    cfg, err := factory.ReadConfig(cliCtx.String("config"))
    if err != nil {
        sigCh <- nil
        return err
    }
    factory.DnfConfig = cfg

    dnf, err := service.NewApp(ctx, cfg)
    if err != nil {
        sigCh <- nil
        return err
    }
    DNF = dnf

    dnf.Start()

    return nil
}
```
As the entrypoint, `action()` initiates a few things:

- Prints the F5GC build information
- Creates a background context which is cancellable:
    ```go=
    ctx, cancel := context.WithCancel(context.Background())
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    ```
    for clean goroutines exit with:
    ```go=
    if err != nil {
        sigCh <- nil
        return err
    }
    ```
    which `cancel()` in a separate dedicated goroutine:
    ```go=
    go func() {
        <-sigCh
        cancel()
    }()
    ```
- Initiates a `factory` which builds the NF in accordance to a `.yaml` config file from `-c` flag.
- Creates a new `DNF` app service which is then started with `dnf.Start()`.

### Logger
Free5GC uses `github.com/sirupsen/logrus` for log formatting. A `logger` package is typically created to format the logs into its submodules. 

`logger/logger.go`
```go=
package logger

import (
    logger_util "github.com/free5gc/util/logger"
    "github.com/sirupsen/logrus"
)

var (
    Log         *logrus.Logger
    NfLog       *logrus.Entry
    MainLog     *logrus.Entry
    InitLog     *logrus.Entry
    CfgLog      *logrus.Entry
    CtxLog      *logrus.Entry
    SBILog      *logrus.Entry
    GinLog      *logrus.Entry
    ConsumerLog *logrus.Entry
    UtilLog     *logrus.Entry
)

func init() {
    fieldsOrder := []string{
        logger_util.FieldNF,
        logger_util.FieldCategory,
    }

    Log = logger_util.New(fieldsOrder)
    NfLog = Log.WithField(logger_util.FieldNF, "DNF")
    MainLog = NfLog.WithField(logger_util.FieldCategory, "Main")
    InitLog = NfLog.WithField(logger_util.FieldCategory, "Init")
    CfgLog = NfLog.WithField(logger_util.FieldCategory, "CFG")
    CtxLog = NfLog.WithField(logger_util.FieldCategory, "CTX")
    SBILog = NfLog.WithField(logger_util.FieldCategory, "SBI")
    GinLog = NfLog.WithField(logger_util.FieldCategory, "GIN")
    ConsumerLog = NfLog.WithField(logger_util.FieldCategory, "Consumer")
    UtilLog = NfLog.WithField(logger_util.FieldCategory, "Util")
}
```

### Factory
A Free5GC utilizes a factory pattern to read configurations from `yaml` files parsed by `gopkg.in/yaml.v2` and is validated with `github.com/asaskevich/govalidator`.

A Free5GC configuration file typically consists of 3 top level components, each are nested structs:
- Info
- Configuration
- Logger

and a read/write mutex to prevent race conditions.

`factory/config.go`
```go=
type Config struct {
    Info          *Info
    Configuration *Configuration
    Logger        *Logger
    sync.RWMutex
}
...
type Configuration struct {
    NfInstanceId    string   `yaml:"nfInstanceId,omitempty" valid:"optional,uuidv4"`
    Sbi             *Sbi     `yaml:"sbi,omitempty" valid:"required"`
    ServiceNameList []string `yaml:"serviceNameList,omitempty" valid:"required"`
    NrfUri          string   `yaml:"nrfUri,omitempty" valid:"url,required"`
    GroupId         string   `yaml:"groupId,omitempty" valid:"type(string),minstringlength(1)"`
}
...
type Sbi struct {
    Scheme       string `yaml:"scheme" valid:"scheme"`
    RegisterIPv4 string `yaml:"registerIPv4,omitempty" valid:"host,required"` // IP that is registered at NRF.
    BindingIPv4  string `yaml:"bindingIPv4,omitempty" valid:"host,required"`  // IP used to run the server in the node.
    Port         int    `yaml:"port,omitempty" valid:"port,required"`
}
```

Each with getter/setter methods that uses the mutex to prevent race condition
```go=
func (c *Config) GetVersion() string {
    c.RWMutex.RLock()
    defer c.RWMutex.RUnlock()

    if c.Info.Version != "" {
        return c.Info.Version
    }
    return ""
}
...
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
...
func (c *Config) GetSbiPort() int {
    c.RLock()
    defer c.RUnlock()
    if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Port != 0 {
        return c.Configuration.Sbi.Port
    }
    return DnfSbiDefaultPort
}
```
Some parses returns default constants
```go=
const (
    DnfDefaultConfigPath = "./NFs/dnf/dnfcfg.yaml"
    DnfSbiDefaultIPv4    = "127.0.0.101"
    DnfSbiDefaultPort    = 8000
    DnfSbiDefaultScheme  = "https"
    DnfDefaultNrfUri     = "https://127.0.0.10:8000"
    DnfDummyUriPrefix    = "/dnf-dummy/v1"
)
```

#### Deserialization
`dnfcfg.yaml`
```yaml=
info:
  version: 0.0.0
  description: DNF sample configuration

configuration:
  sbi:
    scheme: http
    registerIPv4: 127.0.0.101
    bindingIPv4: 127.0.0.101
    port: 8000
  serviceNameList:
    - ndnf-dummy
  nrfUri: http://127.0.0.10:8000

logger:
  enable: true
  level: info
  reportCaller: false
```

The configuration `.yaml` is deserealized with 
`factory/factory.go`
```go=
func InitConfigFactory(f string, cfg *Config) error {
    if f == "" {
        f = DnfDefaultConfigPath
    }

    if content, err := os.ReadFile(f); err != nil {
        return fmt.Errorf("[Factory] %+v", err)
    } else {
        logger.CfgLog.Infof("Read config from [%s]", f)
        if yamlErr := yaml.Unmarshal(content, cfg); yamlErr != nil {
            return fmt.Errorf("[Factory] %+v", yamlErr)
        }
    }

    return nil
}
```

Using Govalidator, the nested struct is validated with nested `validate()` methods.

`factory/config.go`
```go=
func (c *Config) Validate() (bool, error) {
    // Custom struct tag for `valid:"scheme"``
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
...
func (c *Sbi) validate() (bool, error) {
    result, err := govalidator.ValidateStruct(c)
    return result, appendInvalid(err)
}
```
Which is called by the top level function `ReadConfig()`

`factory/factory.go`
```go=
func ReadConfig(cfgPath string) (*Config, error) {
    cfg := &Config{}
    if err := InitConfigFactory(cfgPath, cfg); err != nil {
        return nil, fmt.Errorf("ReadConfig [%s] Error: %+v", cfgPath, err)
    }

    if _, err := cfg.Validate(); err != nil {
        validErrs := err.(govalidator.Errors).Errors()
        for _, validErr := range validErrs {
            logger.CfgLog.Errorf("%+v", validErr)
        }
        logger.CfgLog.Errorf("[-- PLEASE REFER TO SAMPLE CONFIG FILE COMMENTS --]")
        return nil, fmt.Errorf("config validate Error")
    }

    return cfg, nil
}
```

### Service App
#### Structure
A typical NF app consists of the following:
- NF Context
- Config
- Golang Context
- SBI Server
- Consumer
- Processor

represented as a singleton implementing the App interface
`app/app.go`
```go=
type App interface {
    SetLogEnable(enable bool)
    SetLogLevel(level string)
    SetReportCaller(reportCaller bool)

    Start()
    Terminate()

    Context() *dnf_context.DnfContext
    Config() *factory.Config
}
```

`service/init.go`
```go=
var DNF *DnfApp

var _ app.App = &DnfApp{} // Interface checking

type DnfApp struct {
    dnfCtx *dnf_context.DnfContext
    cfg    *factory.Config

    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup

    sbiServer *sbi.Server
    consumer  *consumer.Consumer
    processor *processor.Processor
}
```
each with their own getters with their own names
```go=
func (a *DnfApp) CancelContext() context.Context {
    return a.ctx
}

func (a *DnfApp) Consumer() *consumer.Consumer {
    return a.consumer
}

func (a *DnfApp) Processor() *processor.Processor {
    return a.processor
}
```
initalized by a method `NewApp()`
```go=
func NewApp(ctx context.Context, cfg *factory.Config) (*DnfApp, error) {
    dnf := &DnfApp{
        cfg: cfg,
        wg:  sync.WaitGroup{},
    }
    dnf.SetLogEnable(cfg.GetLogEnable())
    dnf.SetLogLevel(cfg.GetLogLevel())
    dnf.SetReportCaller(cfg.GetLogReportCaller())
    dnf_context.Init()

    processor, err_p := processor.NewProcessor(dnf)
    if err_p != nil {
        return dnf, err_p
    }
    dnf.processor = processor

    consumer, err := consumer.NewConsumer(dnf)
    if err != nil {
        return dnf, err
    }
    dnf.consumer = consumer

    dnf.ctx, dnf.cancel = context.WithCancel(ctx)
    dnf.dnfCtx = dnf_context.GetSelf()

    if dnf.sbiServer, err = sbi.NewServer(dnf); err != nil {
        return nil, err
    }

    DNF = dnf

    return dnf, nil
}
```

that does the following
- Creates an empty struct
- Processes the logger with `logrus` driven methods
- Initializes the NF context
- Initializes the processor
- Initializes the customer
- Creates a cancellable Golang context
- Initializes a new SBI server

#### Start
`service/init.go`
```go=
func (a *DnfApp) Start() {
    logger.InitLog.Infoln("Server started")

    a.wg.Add(1)
    go a.listenShutdownEvent()

    if err := a.sbiServer.Run(context.Background(), &a.wg); err != nil {
        logger.MainLog.Fatalf("Run SBI server failed: %+v", err)
    }

    a.WaitRoutineStopped()
}
```
When the app is started, it starts its SBI server in the background. Alongside it, it creates a goroutine to listen for shutdown events to exit cleanly
```go=
func (a *DnfApp) listenShutdownEvent() {
    defer func() {
        if p := recover(); p != nil {
            logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
        }
        a.wg.Done()
    }()

    <-a.ctx.Done()
    a.terminateProcedure()
}
```
#### Exit
During exit, it deregisters from NRF and stops the SBI server
`service/init.go`
```go=
func (a *DnfApp) terminateProcedure() {
    logger.MainLog.Infof("Terminating DNF...")
    a.CallServerStop()

    // deregister with NRF
    problemDetails, err := a.Consumer().SendDeregisterNFInstance()
    if problemDetails != nil {
        logger.MainLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
    } else if err != nil {
        logger.MainLog.Errorf("Deregister NF instance Error[%+v]", err)
    } else {
        logger.MainLog.Infof("Deregister from NRF successfully")
    }
    logger.MainLog.Infof("CHF SBI Server terminated")
}
```
```go=
func (a *AusfApp) CallServerStop() {
    if a.sbiServer != nil {
        a.sbiServer.Shutdown()
    }
}
```
the function until it finishes. It calls `WaitRoutineStopped()` that waits for the clean exit

```go=
func (a *DnfApp) WaitRoutineStopped() {
    a.wg.Wait()
    logger.MainLog.Infof("DNF App is terminated")
}
```

### Context
`context/context.go`
```go=
type DNFContext struct {
    NfId           string
    SBIPort        int
    RegisterIPv4   string
    BindingIPv4    string
    Url            string
    UriScheme      models.UriScheme
    NrfUri         string
    NrfCertPem     string
    NfService      map[models.ServiceName]models.NrfNfManagementNfService
    OAuth2Required bool
}
```
The context is stored as a struct which stores several important data:
- Network Function ID (Generated UUID or premade)
- API Port
- Register IP Address to be given to NRF
- Binding IP Address the API locally binds to
- URL and its scheme
- NRF URI and its Certificate
- Map of DNF's Services to be registered to the NF

These fields are then initialized:
```go=
var dnfContext DNFContext

func Init() {
    InitDnfContext(&dnfContext)
}
...
func InitDnfContext(context *DNFContext) {
    config := factory.DnfConfig
    logger.InitLog.Infof("dnfconfig Info: Version[%s] Description[%s]\n", config.Info.Version, config.Info.Description)

    configuration := config.Configuration
    sbi := configuration.Sbi

    context.NfId = config.GetNfInstanceId()
    // Defaults
    context.NrfUri = configuration.NrfUri
    context.NrfCertPem = configuration.NrfCertPem
    context.UriScheme = models.UriScheme(configuration.Sbi.Scheme)
    context.RegisterIPv4 = factory.DnfSbiDefaultIPv4
    context.SBIPort = factory.DnfSbiDefaultPort
    if sbi != nil {
        if sbi.RegisterIPv4 != "" {
            context.RegisterIPv4 = sbi.RegisterIPv4
        }
        if sbi.Port != 0 {
            context.SBIPort = sbi.Port
        }

        if sbi.Scheme == "https" {
            context.UriScheme = models.UriScheme_HTTPS
        } else {
            context.UriScheme = models.UriScheme_HTTP
        }

        context.BindingIPv4 = os.Getenv(sbi.BindingIPv4)
        if context.BindingIPv4 != "" {
            logger.InitLog.Info("Parsing ServerIPv4 address from ENV Variable.")
        } else {
            context.BindingIPv4 = sbi.BindingIPv4
            if context.BindingIPv4 == "" {
                logger.InitLog.Warn("Error parsing ServerIPv4 address as string. Using the 0.0.0.0 address as default.")
                context.BindingIPv4 = "0.0.0.0"
            }
        }
    }

    context.Url = string(context.UriScheme) + "://" + context.RegisterIPv4 + ":" + strconv.Itoa(context.SBIPort)

    context.NfService = make(map[models.ServiceName]models.NrfNfManagementNfService)
    AddNfServices(&context.NfService, config, context)
    
    fmt.Println("DNF Context: ", context)
}
```
The data fields are mostly taken from the `factory.DnfConfig` where they are validated and given default values before storing. Additionally, the function `AddNfServices()` is called to fill the service metadata details.
```go=
func AddNfServices(
    serviceMap *map[models.ServiceName]models.NrfNfManagementNfService, config *factory.Config, context *DNFContext,
) {
    var nfService models.NrfNfManagementNfService
    var ipEndPoints []models.IpEndPoint
    var nfServiceVersions []models.NfServiceVersion
    services := *serviceMap

    nfService.ServiceInstanceId = context.NfId
    nfService.ServiceName = models.ServiceName(ServiceName_NDNF_DUMMY)

    var ipEndPoint models.IpEndPoint
    ipEndPoint.Ipv4Address = context.RegisterIPv4
    ipEndPoint.Port = int32(context.SBIPort)
    ipEndPoints = append(ipEndPoints, ipEndPoint)

    var nfServiceVersion models.NfServiceVersion
    nfServiceVersion.ApiFullVersion = config.Info.Version
    nfServiceVersion.ApiVersionInUri = "v1"
    nfServiceVersions = append(nfServiceVersions, nfServiceVersion)

    nfService.Scheme = context.UriScheme
    nfService.NfServiceStatus = models.NfServiceStatus_REGISTERED

    nfService.IpEndPoints = ipEndPoints
    nfService.Versions = nfServiceVersions
    services[ServiceName_NDNF_DUMMY] = nfService
}
```

#### OAuth2 Token Context
```go=
func (c *DNFContext) GetTokenCtx(
    serviceName models.ServiceName, 
    targetNF models.NrfNfManagementNfType,
) (context.Context, *models.ProblemDetails, error) {
    if !c.OAuth2Required {
        return context.TODO(), nil, nil
    }
    return oauth.GetTokenCtx(
        models.NrfNfManagementNfType_AF, 
        targetNF, c.NfId, c.NrfUri, 
        string(serviceName),
    )
}
```
A function is created to return a Golang `context` containing the OAuth2 token via the NRF NF Management. It returns as a `context` for ease of use during API calling.

### SBI Server
`sbi/server.go`
```go=
type ServerDnf interface {
    app.App

    Consumer() *consumer.Consumer
    Processor() *processor.Processor
}

type Server struct {
    ServerDnf

    httpServer *http.Server
    router     *gin.Engine
}
```
An SBI server consists of an `app.App` based interface `ServerDnf` which is implemented by the `Server` struct. It consists of `Consumer()` and `Processor()` which is instead implemented by whatever `app.App` implementation that was initially pased into it in `sbi.NewServer(dnf)`, in this case the `DnfApp` top level implementation in `init.go`, where `consumer.Consumer` and `processor.Processor` are initialized. Additionally, it uses Golang `http.Server` as server and `gin.Engine` as its route handler.
```go=
func NewServer(ausf ServerAusf, tlsKeyLogPath string) (*Server, error) {
    s := &Server{
        ServerAusf: ausf,
    }

    s.router = newRouter(s)

    cfg := s.Config()
    bindAddr := cfg.GetSbiBindingAddr()
    logger.SBILog.Infof("Binding addr: [%s]", bindAddr)
    var err error
    if s.httpServer, err = httpwrapper.NewHttp2Server(bindAddr, tlsKeyLogPath, s.router); err != nil {
        logger.InitLog.Errorf("Initialize HTTP server failed: %v", err)
        return nil, err
    }
    s.httpServer.ErrorLog = log.New(logger.SBILog.WriterLevel(logrus.ErrorLevel), "HTTP2: ", 0)

    return s, nil
}
```
A new server is created with `NewServer()` which creates new router and sets up its binding address based on config. It then uses the `httpwrapper` package from Free5GC to create a new HTTP/2 server.

### Router
`sbi/server.go`
```go=
type ServiceName string

const (
    ServiceName_NDNF_DUMMY ServiceName = "ndnf-dummy"
)
...
func newRouter(s *Server) *gin.Engine {
    router := logger_util.NewGinWithLogrus(logger.GinLog)

    for _, serviceName := range factory.DnfConfig.Configuration.ServiceNameList {
        switch ServiceName(serviceName) {
        case ServiceName_NDNF_DUMMY:
            dnfDummyGroup := router.Group(factory.DnfDummyUriPrefix)
            dnfDummyRoutes := s.getDummyRoutes()
            applyRoutes(dnfDummyGroup, dnfDummyRoutes)

        default:
            logger.SBILog.Warnf("Unsupported service name: %s", serviceName)
        }
    }

    return router
}
```
The function `newRouter()` initializes a `gin.Engine` with route groups depending on the SBI API groups. DNF only has one group `NDNF_DUMMY` which is typecasted into `ServiceName`. 3GPP defined API groups are defined in `github.com/free5gc/openapi/models` package.

`sbi/routes.go`
```go=
package sbi

import "github.com/gin-gonic/gin"

type Route struct {
    Name    string
    Method  string
    Pattern string
    APIFunc gin.HandlerFunc
}

func applyRoutes(group *gin.RouterGroup, routes []Route) {
    for _, route := range routes {
        switch route.Method {
        case "GET":
            group.GET(route.Pattern, route.APIFunc)
        case "POST":
            group.POST(route.Pattern, route.APIFunc)
        case "PUT":
            group.PUT(route.Pattern, route.APIFunc)
        case "PATCH":
            group.PATCH(route.Pattern, route.APIFunc)
        case "DELETE":
            group.DELETE(route.Pattern, route.APIFunc)
        }
    }
}
```
Every F5GC NF SBI contains `route.go` which abstracts API route creation into a `Route` struct which defines each endpoint as a data group declaration.
`sbi/api_dummy.go`
```go=
func (s *Server) getDummyRoutes() []Route {
    return []Route{
        {
            Name:    "Index",
            Method:  http.MethodGet,
            Pattern: "/",
            APIFunc: s.HTTPDummyMessage,
        },
    }
}

func (s *Server) HTTPDummyMessage(c *gin.Context) {
    c.String(http.StatusOK, "Hello DNF!")
}
```
During initial run, the SBI Server registers itself via the `Consumer` to the NRF then starts up the Golang HTTP server.

`sbi/server.go`
```go=
func (s *Server) Run(traceCtx context.Context, wg *sync.WaitGroup) error {
    var err error
    _, s.Context().NfId, err = s.Consumer().RegisterNFInstance(context.Background())
    if err != nil {
        logger.InitLog.Errorf("DNF register to NRF Error[%s]", err.Error())
    }

    wg.Add(1)
    go s.startServer(wg)

    return nil
}
...
func (s *Server) startServer(wg *sync.WaitGroup) {
    defer func() {
        if p := recover(); p != nil {
            // Print stack for panic to log. Fatalf() will let program exit.
            logger.SBILog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
            s.Terminate()
        }
        wg.Done()
    }()

    logger.SBILog.Infof("Start SBI server (listen on %s)", s.httpServer.Addr)

    var err error
    err = s.httpServer.ListenAndServe()

    if err != nil && err != http.ErrServerClosed {
        logger.SBILog.Errorf("SBI server error: %v", err)
    }
    logger.SBILog.Infof("SBI server (listen on %s) stopped", s.httpServer.Addr)
}
```

### Consumer
The consumer utilizes other NFs' SBI interfaces. It consists of the respective NFs' that are interracted with, each with the API groups that the NF would use.
DNF would require the following:
- NRF's NFManagement to register itself (a lot of NFs use this)
- NRF's NFDiscovery to discover other NFs (and maybe itself)
- UDM's SubscriberDataManagement to fetch users' data

`sbi/consumer/consumer.go`
```go=
type ConsumerDnf interface {
    app.App
}

type Consumer struct {
    ConsumerDnf

    *nnrfService
    *nudmService
}
```

`sbi/consumer/nrf_service.go`
```go=
import (
    ...
    Nnrf_NFDiscovery "github.com/free5gc/openapi/nrf/NFDiscovery"
    Nnrf_NFManagement "github.com/free5gc/openapi/nrf/NFManagement"
)

type nnrfService struct {
    consumer *Consumer

    nfMngmntMu sync.RWMutex
    nfDiscMu   sync.RWMutex

    nfMngmntClients map[string]*Nnrf_NFManagement.APIClient
    nfDiscClients   map[string]*Nnrf_NFDiscovery.APIClient
}
```

`sbi/consumer/udm_service.go`
```go=
import (
    ...
    Nudm_SDM "github.com/free5gc/openapi/udm/SubscriberDataManagement"
)

type nudmService struct {
    consumer *Consumer

    sdmMu sync.RWMutex

    sdmClients map[string]*Nudm_SDM.APIClient
}
```

Each NF service contains a RW mutex to prevent collisions and an API client dictionary to store multiple connections at the same time, each initialized when the `Consumer` is initialized. Each service also hold an instance of the top level `Consumer` for two-way coupling.

`sbi/consumer/consumer.go`
```go=
func NewConsumer(dnf ConsumerDnf) (*Consumer, error) {
    c := &Consumer{
        ConsumerDnf: dnf,
    }

    c.nnrfService = &nnrfService{
        consumer:        c,
        nfMngmntClients: make(map[string]*Nnrf_NFManagement.APIClient),
    }

    c.nudmService = &nudmService{
        consumer:   c,
        sdmClients: make(map[string]*Nudm_SDM.APIClient),
    }

    return c, nil
}
```

#### NRF Service

##### NF Management Service
`consumer/nrf_service.go`
```go=
func (s *nnrfService) getNFManagementClient(uri string) *Nnrf_NFManagement.APIClient {
    if uri == "" {
        return nil
    }
    s.nfMngmntMu.RLock()
    client, ok := s.nfMngmntClients[uri]
    if ok {
        s.nfMngmntMu.RUnlock()
        return client
    }

    configuration := Nnrf_NFManagement.NewConfiguration()
    configuration.SetBasePath(uri)
    client = Nnrf_NFManagement.NewAPIClient(configuration)

    s.nfMngmntMu.RUnlock()
    s.nfMngmntMu.Lock()
    defer s.nfMngmntMu.Unlock()
    s.nfMngmntClients[uri] = client
    return client
}
```
Before calling its service, `getNFManagementClient()` is called to retrieve the appropriate API client for NF Management service from an NRF URI. It locks its mutex, checks whether a client was previously initialized and reuse it or reinitiates a new one if it hadn't. It uses the `NewConfiguration()` method where it sets its base URI path to then create a new client with `NewAPIClient()`. The mutex is set to write lock before adding the previously made client into the client map `nfMngmntClients[]`.

###### Register NF Instance
```go=
func (s *nnrfService) RegisterNFInstance(ctx context.Context) (
    string, string, error,
) {
```
Function to add the NF into NRF's registry. It issues a `PUT` command then returns the resource URI and the NF's retrieval ID.
```go=
...
    dnfContext := s.consumer.Context()
    client := s.getNFManagementClient(dnfContext.NrfUri)
    nfProfile, err := s.buildNfProfile(dnfContext)
    if err != nil {
        return "", "", errors.Wrap(err, "RegisterNFInstance buildNfProfile()")
    }
...
}

func (s *nnrfService) buildNfProfile(dnfContext *dnf_context.DNFContext) (
    models.NrfNfManagementNfProfile, error,
) {
    var profile models.NrfNfManagementNfProfile
    profile.NfInstanceId = dnfContext.NfId
    profile.NfType = models.NrfNfManagementNfType_AF
    profile.NfStatus = models.NrfNfManagementNfStatus_REGISTERED
    profile.Ipv4Addresses = append(profile.Ipv4Addresses, dnfContext.RegisterIPv4)

    services := []models.NrfNfManagementNfService{}
    for _, nfService := range dnfContext.NfService {
        services = append(services, nfService)
    }
    if len(services) > 0 {
        profile.NfServices = services
    }

    return profile, nil
}
```
RegisterNFInstance accepts an NF ID and an NFProfile defined as a schema. To build NFProfile, `buildNfProfile()` is called which initializes a `models.NrfNfManagementNfProfile` with the following:
- NF Instance ID
- NF Type (AF in this case, to not modify the NRF)
- NF Status (Registered)
- IP Address(es) set to `RegisterIPv4`
- Every single available services this NF provided (retrieved from its NF context)

```go=
func (s *nnrfService) RegisterNFInstance(ctx context.Context) (
    string, string, error,
) {
...
    var nf models.NrfNfManagementNfProfile
    var res *Nnrf_NFManagement.RegisterNFInstanceResponse
    registerNFInstanceRequest := &Nnrf_NFManagement.RegisterNFInstanceRequest{
        NfInstanceID:             &dnfContext.NfId,
        NrfNfManagementNfProfile: &nfProfile,
    }

    var resourceNrfUri string
    var retrieveNfInstanceID string
    for {
        ...
        res, err = client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, registerNFInstanceRequest)
        if err != nil || res == nil {
            logger.ConsumerLog.Errorf("DNF register to NRF Error[%v]", err)
            if apiErr, ok := err.(openapi.GenericOpenAPIError); ok {
                if regErr, ok := apiErr.Model().(Nnrf_NFManagement.RegisterNFInstanceError); ok {
                    logger.ConsumerLog.Errorf("%v", regErr.ProblemDetails.Detail)
                }
            }
            time.Sleep(2 * time.Second)
            continue
        }
...
```
The previously built NFProfile is added into `RegisterNFInstanceRequest` struct alongside the previously mentioned `NfInstanceID`.

Through a indefinite loop, it sends the request through the client's method where it checks for error. If an error is found, it delays itself before indefinitely trying again.

```go=
...
        nf = res.NrfNfManagementNfProfile
        if res.Location == "" {
            break
        } else {
            resourceUri := res.Location
            resourceNrfUri = resourceUri[:strings.Index(resourceUri, "/nnrf-nfm/")]
            retrieveNfInstanceID = resourceUri[strings.LastIndex(resourceUri, "/")+1:]

            oauth2 := false
            if nf.CustomInfo != nil {
                v, ok := nf.CustomInfo["oauth2"].(bool)
                if ok {
                    oauth2 = v
                    logger.MainLog.Infoln("OAuth2 setting receive from NRF: ", oauth2)
                }
            }
            dnf_context.GetSelf().OAuth2Required = oauth2
            if oauth2 && dnf_context.GetSelf().NrfCertPem == "" {
                logger.CfgLog.Error("OAuth2 enable but no nrfCertPem provided in config.")
            }

            break
        }
    }
    return resourceNrfUri, retrieveNfInstanceID, err
}
```
Once the request goes through, the resource URI and retrieve NF instance ID is extracted from the Location Header. Additionally, the NRF's OAuth2 configuration is gained and stored into the NF's context to handle future processing.

![image](https://hackmd.io/_uploads/r17yAZLEMx.png)

###### Deregister NF Instance
```go=
func (s *nnrfService) SendDeregisterNFInstance() (*models.ProblemDetails, error) {
    logger.ConsumerLog.Infof("[DNF] Send Deregister NFInstance")

    ctx, problemDetail, err := dnf_context.GetSelf().GetTokenCtx(models.ServiceName_NNRF_NFM, models.NrfNfManagementNfType_NRF)
    if err != nil {
        return problemDetail, err
    }
...    
```
OAuth2 requires every requests to check for its token first. `GetTokenCtx` is called to retrieve an appropriate context for the API call.

```go=
...
    dnfContext := s.consumer.Context()
    client := s.getNFManagementClient(dnfContext.NrfUri)
    deregisterNFInstanceRequest := &Nnrf_NFManagement.DeregisterNFInstanceRequest{
        NfInstanceID: &dnfContext.NfId,
    }

    _, err = client.NFInstanceIDDocumentApi.DeregisterNFInstance(ctx, deregisterNFInstanceRequest)
    ...
}
```
Similar to `RegisterNFInstance()`, a struct `DeregisterNFInstanceRequest` is used to form the request alongside the OAuth2-containing context. NF Deregistration only requires the NF's ID.

![image](https://hackmd.io/_uploads/r1YBkGt4Me.png)


##### NF Discovery Service
```go=
func (s *nnrfService) getNFDiscClient(uri string) *Nnrf_NFDiscovery.APIClient {
    if uri == "" {
        return nil
    }
    s.nfDiscMu.RLock()
    client, ok := s.nfDiscClients[uri]
    if ok {
        s.nfDiscMu.RUnlock()
        return client
    }

    configuration := Nnrf_NFDiscovery.NewConfiguration()
    configuration.SetBasePath(uri)
    client = Nnrf_NFDiscovery.NewAPIClient(configuration)

    s.nfDiscMu.RUnlock()
    s.nfDiscMu.Lock()
    defer s.nfDiscMu.Unlock()
    s.nfDiscClients[uri] = client
    return client
}
```
Getting a discovery service has the exact same flow as management service.

###### Search NF Instance
```go=
func (s *nnrfService) SendSearchNFInstances(nrfUri string, targetNfType, requestNfType models.NrfNfManagementNfType) (*models.SearchResult, error) {
    // Set client and set url
    searchNFRequest := Nnrf_NFDiscovery.SearchNFInstancesRequest{
        TargetNfType:    &targetNfType,
        RequesterNfType: &requestNfType,
    }
    ...
```
Create request struct with the following:
- Target NF Type
- Requester NF Type
```go=
    ...
    client := s.getNFDiscClient(nrfUri)
    if client == nil {
        return nil, openapi.ReportError("nrf not found")
    }

    ctx, _, err := dnf_context.GetSelf().GetTokenCtx(models.ServiceName_NNRF_DISC, models.NrfNfManagementNfType_NRF)
    if err != nil {
        return nil, err
    }
    res, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, &searchNFRequest)
    ...
```
Creates client and sends request.
```go=
    ...
    var result *models.SearchResult
    if err != nil {
        logger.ConsumerLog.Errorf("SearchNFInstances failed: %+v", err)
    }
    if res != nil {
        result = &res.SearchResult
        logger.ConsumerLog.Infof("Found NF Instance: %v", result.NfInstances[0].NfInstanceId)
    }
    return result, err
}
```
Handle error and process result.

#### UDM Service

##### Subscriber Data Management Service
`consumer/udm_service.go`
```go=
func (s *nudmService) getSubscriberDMngmntClients(uri string) *Nudm_SDM.APIClient {
    if uri == "" {
        return nil
    }
    s.sdmMu.RLock()

    client, ok := s.sdmClients[uri]
    if ok {
        return client
    }

    configuration := Nudm_SDM.NewConfiguration()
    configuration.SetBasePath(uri)
    client = Nudm_SDM.NewAPIClient(configuration)

    s.sdmMu.RUnlock()
    s.sdmMu.Lock()
    defer s.sdmMu.Unlock()
    s.sdmClients[uri] = client
    return client
}
```
Getting a UDM Service client has the exact same flow as getting an NRF Service client.

###### Slice Selection Subscription Data Retrieval
An SBI interraction is similarly done as previous.
```go=
func (s *nudmService) GetNSSAI(supi string, mcc string, mnc string) (
    *models.ProblemDetails,
    error,
) {
    dnfContext := s.consumer.Context()
    client := s.getSubscriberDMngmntClients(dnfContext.UdmUri)

    getNSSAIRequest := Nudm_SDM.GetNSSAIRequest{
        Supi: &supi,
        PlmnId: &models.PlmnId{
            Mcc: mcc,
            Mnc: mnc,
        },
    }
    ...
    
```
A client is get based on the UDM URI and the request struct is filled with the appropriate data:
- SUPI
- PLMN ID
```go=
    ...
    ctx, problemDetails, err := dnf_context.GetSelf().GetTokenCtx(models.ServiceName_NUDM_SDM, models.NrfNfManagementNfType_UDM)
    if err != nil {
        return problemDetails, err
    }
    res, err := client.SliceSelectionSubscriptionDataRetrievalApi.GetNSSAI(ctx, &getNSSAIRequest)
    ...
```
An OAuth2 context is get and the API request is sent based on the OpenAPI method
```go=
    ...
    if err != nil {
        if apiErr, ok := err.(openapi.GenericOpenAPIError); ok {
            if errModel, ok := apiErr.Model().(Nudm_SDM.GetNSSAIError); ok {
                problemDetails = &errModel.ProblemDetails
            } else {
                err = openapi.ReportError("openapi error")
            }
        } else {
            return nil, openapi.ReportError("openapi error")
        }
    }

    logger.ConsumerLog.Infof("NSSAI of %s: %v", supi, res.Nssai.SingleNssais[0])

    return problemDetails, err
}
```
The error/data is handled.

### Processor
To create a new endpoint, a new route Dummy Process is created
`sbi/routes.go`
```go=
...
{
    Name:    "Dummy Process",
    Method:  http.MethodGet,
    Pattern: "/dummy",
    APIFunc: s.HTTPDummyProcess,
},
...
```
Since it doesnt receive any request body, it jumps straight to the handler
```go=
func (s *Server) HTTPDummyProcess(c *gin.Context) {
    s.Processor().HandleDummyProcess(c)
}
```
which jumps straight to its process function after logging it
`processor/dummy.go`
```go=
func (p *Processor) HandleDummyProcess(c *gin.Context) {
    logger.DummyLog.Infof("DUMMY PROCESSING YEAH!!!!")

    p.DummyProcess(c)
}
```

#### Dummy Process
The Dummy Process has several objectives:
- Discover itself, the DNF
- Finds the NSSAI of an UE described in the config file.
- Returns all the resulting data as a json.

```go=
func (p *Processor) DummyProcess(c *gin.Context) {
    dnfContext := dnf_context.GetSelf()
    nrfUri := dnfContext.NrfUri
    ...
```
It first gets its own context
```go=
    ...
    targetNfType := models.NrfNfManagementNfType_AF
    requestNfType := models.NrfNfManagementNfType_AF

    searchResult, err := p.Consumer().SendSearchNFInstances(nrfUri, targetNfType, requestNfType)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
        })
    }
    ...
```
It searches for itself, which was registered as an AF.
```go=
    ...
    nssai, err := p.Consumer().GetNSSAI(dnfContext.SearchSupi, dnfContext.SearchMCC, dnfContext.SearchMNC)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
        })
    }
    ...
```
It uses `GetNSSAI` by using the SUPI stored in the config file which was stored into the NF Context.
```go=

    // Return as JSON

    c.JSON(http.StatusOK, gin.H{
        "searchResult": searchResult,
        "nssai":        nssai,
    })
}
```
It returns its findings as JSON like a regular API.

## Testing
### Hello DNF
![image](https://hackmd.io/_uploads/HJ1UhbKNGe.png)

### Dummy Process