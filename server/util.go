package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cosmossdk.io/log"
	tmcfg "github.com/cometbft/cometbft/config"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	tmlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	serverlog "github.com/cosmos/cosmos-sdk/server/log"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	sagalog "github.com/sagaxyz/saga-sdk/log"
)

const FlagJsonLogFile = "json-log-file"

func bindFlags(basename string, cmd *cobra.Command, v *viper.Viper) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("bindFlags failed: %v", r)
		}
	}()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		err = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", basename, strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))))
		if err != nil {
			panic(err)
		}

		err = v.BindPFlag(f.Name, f)
		if err != nil {
			panic(err)
		}

		// Apply the viper config value to the flag when the flag is not set and
		// viper has a value.
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				panic(err)
			}
		}
	})

	return err
}

// InterceptConfigsPreRunHandler performs a pre-run function for the root daemon
// application command. It will create a Viper literal and a default server
// Context. The server Tendermint configuration will either be read and parsed
// or created and saved to disk, where the server Context is updated to reflect
// the Tendermint configuration. It takes custom app config template and config
// settings to create a custom Tendermint configuration. If the custom template
// is empty, it uses default-template provided by the server. The Viper literal
// is used to read and parse the application configuration. Command handlers can
// fetch the server Context to get the Tendermint configuration or to get access
// to Viper.
func InterceptConfigsPreRunHandler(cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}, tmConfig *tmcfg.Config) error {
	serverCtx := sdkserver.NewDefaultContext()

	// Get the executable name and configure the viper instance so that environmental
	// variables are checked based off that name. The underscore character is used
	// as a separator.
	executableName, err := os.Executable()
	if err != nil {
		return err
	}

	basename := path.Base(executableName)

	// configure the viper instance
	if err := serverCtx.Viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	if err := serverCtx.Viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return err
	}

	serverCtx.Viper.SetEnvPrefix(basename)
	serverCtx.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	serverCtx.Viper.AutomaticEnv()

	// intercept configuration files, using both Viper instances separately
	config, err := interceptConfigs(serverCtx.Viper, customAppConfigTemplate, customAppConfig, tmConfig)
	if err != nil {
		return err
	}

	// return value is a tendermint configuration object
	serverCtx.Config = config
	if err = bindFlags(basename, cmd, serverCtx.Viper); err != nil {
		return err
	}

	var opts []log.Option
	if serverCtx.Viper.GetString(flags.FlagLogFormat) == tmcfg.LogFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}

	// check and set filter level or keys for the logger if any
	logLvlStr := serverCtx.Viper.GetString(flags.FlagLogLevel)
	if logLvlStr != "" {
		logLvl, err := zerolog.ParseLevel(logLvlStr)
		switch {
		case err != nil:
			// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
			filterFunc, err := log.ParseLogLevel(logLvlStr)
			if err != nil {
				return err
			}

			opts = append(opts, log.FilterOption(filterFunc))
		case serverCtx.Viper.GetBool(tmcli.TraceFlag):
			// Check if the CometBFT flag for trace logging is set if it is then setup a tracing logger in this app as well.
			// Note it overrides log level passed in `log_levels`.
			opts = append(opts, log.LevelOption(zerolog.TraceLevel))
		default:
			opts = append(opts, log.LevelOption(logLvl))
		}
	}

	var logFile io.Writer
	jsonLogFile := serverCtx.Viper.GetString(FlagJsonLogFile)
	if jsonLogFile != "" {
		logFile, err = os.OpenFile(
			jsonLogFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0660,
		)
		if err != nil {
			return err
		}
	}
	logger := sagalog.NewLogger(tmlog.NewSyncWriter(os.Stdout), logFile, opts...)

	serverCtx.Logger = serverlog.CometLoggerWrapper{
		Logger: logger,
	}
	return sdkserver.SetCmdServerContext(cmd, serverCtx)
}

// interceptConfigs parses and updates a Tendermint configuration file or
// creates a new one and saves it. It also parses and saves the application
// configuration file. The Tendermint configuration file is parsed given a root
// Viper object, whereas the application is parsed with the private package-aware
// viperCfg object.
func interceptConfigs(rootViper *viper.Viper, customAppTemplate string, customConfig interface{}, tmConfig *tmcfg.Config) (*tmcfg.Config, error) {
	rootDir := rootViper.GetString(flags.FlagHome)
	configPath := filepath.Join(rootDir, "config")
	tmCfgFile := filepath.Join(configPath, "config.toml")

	conf := tmConfig

	switch _, err := os.Stat(tmCfgFile); {
	case os.IsNotExist(err):
		tmcfg.EnsureRoot(rootDir)

		if err = conf.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("error in config file: %w", err)
		}

		defaultCometCfg := tmcfg.DefaultConfig()
		// The SDK is opinionated about those comet values, so we set them here.
		// We verify first that the user has not changed them for not overriding them.
		if conf.Consensus.TimeoutCommit == defaultCometCfg.Consensus.TimeoutCommit {
			conf.Consensus.TimeoutCommit = 5 * time.Second
		}
		if conf.RPC.PprofListenAddress == defaultCometCfg.RPC.PprofListenAddress {
			conf.RPC.PprofListenAddress = "localhost:6060"
		}
		tmcfg.WriteConfigFile(tmCfgFile, conf)
	case err != nil:
		return nil, err

	default:
		rootViper.SetConfigType("toml")
		rootViper.SetConfigName("config")
		rootViper.AddConfigPath(configPath)

		if err := rootViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read in %s: %w", tmCfgFile, err)
		}
	}

	// Read into the configuration whatever data the viper instance has for it.
	// This may come from the configuration file above but also any of the other
	// sources viper uses.
	if err := rootViper.Unmarshal(conf); err != nil {
		return nil, err
	}

	conf.SetRoot(rootDir)

	appCfgFilePath := filepath.Join(configPath, "app.toml")
	if _, err := os.Stat(appCfgFilePath); os.IsNotExist(err) {
		if customAppTemplate != "" {
			config.SetConfigTemplate(customAppTemplate)

			if err = rootViper.Unmarshal(&customConfig); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", appCfgFilePath, err)
			}

			config.WriteConfigFile(appCfgFilePath, customConfig)
		} else {
			appConf, err := config.ParseConfig(rootViper)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", appCfgFilePath, err)
			}

			config.WriteConfigFile(appCfgFilePath, appConf)
		}
	}

	rootViper.SetConfigType("toml")
	rootViper.SetConfigName("app")
	rootViper.AddConfigPath(configPath)

	if err := rootViper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge configuration: %w", err)
	}

	return conf, nil
}
