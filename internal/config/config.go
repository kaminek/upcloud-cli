package config

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	debug "github.com/UpCloudLtd/cli/internal/log"
	internal "github.com/UpCloudLtd/cli/internal/service"
	"github.com/UpCloudLtd/cli/internal/terminal"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/adrg/xdg"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// KeyClientTimeout defines the viper configuration key used to define client timeout
	KeyClientTimeout = "client-timeout"
	// KeyOutput defines the viper configuration key used to define the output
	KeyOutput = "output"
	// ValueOutputHuman defines the viper configuration value used to define human-readable output
	ValueOutputHuman = "human"
	// ValueOutputYAML defines the viper configuration value used to define YAML output
	ValueOutputYAML = "yaml"
	// ValueOutputJSON defines the viper configuration value used to define JSON output
	ValueOutputJSON = "json"

	// env vars custom prefix
	envPrefix = "UPCLOUD"
)

var (
	// Version contains the current version.
	Version = "dev"
	// BuildDate contains a string with the build date.
	BuildDate = "unknown"
)

// New returns a new instance of Config bound to the given viper instance
func New() *Config {
	return &Config{viper: viper.New()}
}

// GlobalFlags holds information on the flags shared among all commands
type GlobalFlags struct {
	ConfigFile    string        `valid:"-"`
	ClientTimeout time.Duration `valid:"-"`
	Colors        bool          `valid:"-"`
	Debug         bool          `valid:"-"`
	OutputFormat  string        `valid:"in(human|json|yaml)"`
	Wait          bool          `valid:"-"`
}

// Config holds the configuration for running upctl
type Config struct {
	viper   *viper.Viper
	ns      string
	flagSet *pflag.FlagSet
	// TODO: remove this after refactored
	Service     internal.Wrapper
	GlobalFlags GlobalFlags
}

// Load loads config and sets up service
func (s *Config) Load() error {
	v := s.Viper()

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()
	v.SetConfigName("upctl")
	v.SetConfigType("yaml")

	configFile := s.GlobalFlags.ConfigFile
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Support XDG default config home dir and common config dirs
		v.AddConfigPath(xdg.ConfigHome)
		v.AddConfigPath("$HOME/.config") // for MacOS as XDG config is not common
	}

	// Attempt to read the config file, ignoring only config file not found errors
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.Set("config", v.ConfigFileUsed())

	if s.GlobalFlags.Debug {
		debug.Debug("Load cli", "User config after loading and setting up service")
		v.Debug()
	}
	return nil

}

// Viper returns a reference to the viper instance
func (s *Config) Viper() *viper.Viper {
	return s.viper
}

// SetNamespace sets the configuration namespace
func (s *Config) SetNamespace(ns string) {
	s.ns = ns
}

// IsSet return true if the key is set in the current namespace
func (s *Config) IsSet(key string) bool {
	return s.viper.IsSet(s.prependNs(key))
}

// Get return the value of the key in the current namespace
func (s *Config) Get(key string) interface{} {
	return s.viper.Get(s.prependNs(key))
}

// GetString is a convenience method of getting a configuration value in the current namespace as a string
func (s *Config) GetString(key string) string {
	return s.viper.GetString(s.prependNs(key))
}

// Top returns a *copy* of the Config with no namespace
func (s *Config) Top() *Config {
	clone := *s
	clone.SetNamespace("")
	return &clone
}

// FlagByKey returns pflag.Flag associated with a key in config
func (s *Config) FlagByKey(key string) *pflag.Flag {
	if s.flagSet == nil {
		s.flagSet = &pflag.FlagSet{}
	}
	return s.flagSet.Lookup(key)
}

// BoundFlags returns the list of all the flags given to the config
func (s *Config) BoundFlags() []*pflag.Flag {
	if s.flagSet == nil {
		s.flagSet = &pflag.FlagSet{}
	}
	var r []*pflag.Flag
	s.flagSet.VisitAll(func(flag *pflag.Flag) {
		r = append(r, flag)
	})
	return r
}

// ConfigBindFlagSet sets the config flag set and binds them to the viper instance
func (s *Config) ConfigBindFlagSet(flags *pflag.FlagSet) {
	if flags == nil {
		panic("Nil flagset")
	}
	flags.VisitAll(func(flag *pflag.Flag) {
		_ = s.viper.BindPFlag(s.prependNs(flag.Name), flag)
		// s.flagSet.AddFlag(flag)
	})
}

func (s *Config) prependNs(key string) string {
	if s.ns == "" {
		return key
	}
	return fmt.Sprintf("%s.%s", s.ns, key)
}

// Commonly used keys as accessors

// Output is a convenience method for getting the user specified output
func (s *Config) Output() string {
	return s.viper.GetString(KeyOutput)
}

// OutputHuman is a convenience method that returns true if the user specified human-readable output
func (s *Config) OutputHuman() bool {
	return s.Output() == ValueOutputHuman
}

// ClientTimeout is a convenience method that returns the user specified client timeout
func (s *Config) ClientTimeout() time.Duration {
	return s.viper.GetDuration(KeyClientTimeout)
}

// InteractiveUI is a convenience method that returns true if the user has requested human output and the terminal supports it.
func (s *Config) InteractiveUI() bool {
	return terminal.IsStdoutTerminal() && s.OutputHuman()
}

type transport struct{}

// RoundTrip implements http.RoundTripper
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return cleanhttp.DefaultTransport().RoundTrip(req)
}

// CreateService creates a new service instance and puts in the conf struct
func (s *Config) CreateService() (internal.AllServices, error) {
	username := s.Top().GetString("username")
	password := s.Top().GetString("password")

	if username == "" || password == "" {
		// nb. this might give silghtly unexpected results on OS X, as xdg.ConfigHome points to ~/Library/Application Support
		// while we really use/prefer/document ~/.config - which does work on osx as well but won't be displayed here.
		// TODO: fix this?
		configDetails := fmt.Sprintf("default location %s", filepath.Join(xdg.ConfigHome, "upctl.yaml"))
		if s.Top().GetString("config") != "" {
			configDetails = fmt.Sprintf("used %s", s.Top().GetString("config"))
		}
		return nil, fmt.Errorf("user credentials not found, these must be set in config file (%s) or via environment variables", configDetails)
	}

	hc := &http.Client{Transport: &transport{}}
	hc.Timeout = s.ClientTimeout()

	whc := client.NewWithHTTPClient(
		username,
		password,
		hc,
	)
	whc.UserAgent = fmt.Sprintf("upctl/%s", Version)
	svc := service.New(whc)
	// TODO; remove this when refactor is complete
	s.Service = internal.Wrapper{Service: svc}
	return svc, nil
}
