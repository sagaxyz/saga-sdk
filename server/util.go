package server

import (
	"io"
	"os"

	"cosmossdk.io/log"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	sagalog "github.com/sagaxyz/saga-sdk/log"
)

const FlagJsonLogFile = "json-log-file"

func InterceptConfigsPreRunHandler(cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}, cmtConfig *cmtcfg.Config) error {
	serverCtx, err := server.InterceptConfigsAndCreateContext(cmd, customAppConfigTemplate, customAppConfig, cmtConfig)
	if err != nil {
		return err
	}

	// overwrite default server logger
	logger, err := CreateSDKLogger(serverCtx, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	serverCtx.Logger = logger.With(log.ModuleKey, "server")

	// set server context
	return server.SetCmdServerContext(cmd, serverCtx)
}

func CreateSDKLogger(ctx *server.Context, out io.Writer) (logger log.Logger, err error) {
	var opts []log.Option
	if ctx.Viper.GetString(flags.FlagLogFormat) == flags.OutputFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}
	opts = append(opts,
		log.ColorOption(!ctx.Viper.GetBool(flags.FlagLogNoColor)),
		// We use CometBFT flag (cmtcli.TraceFlag) for trace logging.
		log.TraceOption(ctx.Viper.GetBool(server.FlagTrace)))

	// check and set filter level or keys for the logger if any
	logLvlStr := ctx.Viper.GetString(flags.FlagLogLevel)
	if logLvlStr == "" {
		return log.NewLogger(out, opts...), nil
	}

	logLvl, err := zerolog.ParseLevel(logLvlStr)
	switch {
	case err != nil:
		// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
		filterFunc, err := log.ParseLogLevel(logLvlStr)
		if err != nil {
			return nil, err
		}

		opts = append(opts, log.FilterOption(filterFunc))
	default:
		opts = append(opts, log.LevelOption(logLvl))
	}

	var logFile io.Writer
	jsonLogFile := ctx.Viper.GetString(FlagJsonLogFile)
	if jsonLogFile != "" {
		logFile, err = os.OpenFile(
			jsonLogFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0660,
		)
		if err != nil {
			return
		}
	}

	logger = sagalog.NewLogger(out, logFile, opts...)
	return
}
