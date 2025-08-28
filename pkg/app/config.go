package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/costa92/go-protoc/v2/pkg/version"
)

const ConfigFlagName = "config"

var CfgFile string

// AddConfigFlag adds flags for a specific server to the specified FlagSet object.
// It also sets a passed functions to read values from configuration file into viper
// when each cobra command's Execute method is called.
func AddConfigFlag(fs *pflag.FlagSet, name string, watch bool) {
	// 添加配置文件标志
	fs.StringVarP(&CfgFile, ConfigFlagName, "c", CfgFile, "Read configuration from specified `FILE`, "+
		"support JSON, TOML, YAML, HCL, or Java properties formats.")

	// 只有在logger已经初始化的情况下才输出日志
	if logger.GetDefaultLogger() != nil {
		logger.Infow("Adding configuration flag", "name", name, "watch", watch, "cfgFile", CfgFile)
	}
	// Enable viper's automatic environment variable parsing. This means
	// that viper will automatically read values corresponding to viper
	// variables from environment variables.
	viper.AutomaticEnv()
	// Set the environment variable prefix. Use the strings.ReplaceAll function
	// to replace hyphens with underscores in the name, and use strings.ToUpper
	// to convert the name to uppercase, then set it as the prefix for environment variables.
	viper.SetEnvPrefix(strings.ReplaceAll(strings.ToUpper(name), "-", "_"))
	// Set the replacement rules for environment variable keys. Use the
	// strings.NewReplacer function to specify replacing periods and hyphens with underscores.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cobra.OnInitialize(func() {
		logger.Infow("Reading configuration file", "name", name, "cfgFile", CfgFile)
		if CfgFile != "" {
			viper.SetConfigFile(CfgFile)
		} else {
			viper.AddConfigPath(".")
			viper.AddConfigPath("configs")

			if names := strings.Split(name, "-"); len(names) > 1 {
				viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0]))
				viper.AddConfigPath(filepath.Join("/etc", names[0]))
			}

			viper.SetConfigName(name)
		}

		logger.Debugw("Reading configuration file", "file", CfgFile)

		if err := viper.ReadInConfig(); err != nil {
			logger.Errorw("Failed to read configuration file", "error", err, "file", CfgFile)
		} else {
			logger.Infow("Success to read configuration file", "file", viper.ConfigFileUsed())
			
			// 配置文件读取成功后，立即重新初始化日志器以应用日志配置
			reinitializeLoggerFromConfig()
		}

		if watch {
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				logger.Debugw("Config file changed", "name", e.Name)
				// 配置变更时重新初始化日志器
				reinitializeLoggerFromConfig()
			})
		}
	})
}

// reinitializeLoggerFromConfig 根据配置文件重新初始化日志器
func reinitializeLoggerFromConfig() {
	var logOptions *logger.LogsOptions

	// 优先使用简化配置（新方式）
	if viper.IsSet("log.preset") {
		quickConfig := parseQuickConfig()
		logOptions = quickConfig.ToFullOptions()
		logger.Infow("Using simplified log configuration", "preset", quickConfig.Preset, "type", quickConfig.Type)
	} else {
		// 兼容旧的详细配置
		logOptions = parseDetailedConfig()
		logger.Infow("Using detailed log configuration", "type", logOptions.Type)
	}

	// 添加版本信息到初始字段
	addVersionInfoToLogger(logOptions)

	// 创建并设置全局日志器
	if globalLogger, err := logger.NewLogger(logOptions); err != nil {
		logger.Errorw("Failed to reinitialize logger from config", "error", err)
	} else {
		logger.SetDefaultLogger(globalLogger)
		versionInfo := version.Get()
		logger.Infow("Logger reinitialized from configuration file", 
			"level", logOptions.Level, 
			"format", logOptions.Format,
			"service", versionInfo.ServiceName,
			"type", string(logOptions.Type))
	}
}

// parseQuickConfig 解析简化配置
func parseQuickConfig() *logger.QuickConfig {
	config := logger.DefaultQuickConfig()
	
	if viper.IsSet("log.preset") {
		config.Preset = logger.LogPreset(viper.GetString("log.preset"))
	}
	if viper.IsSet("log.type") {
		config.Type = viper.GetString("log.type")
	}
	if viper.IsSet("log.level") {
		config.Level = viper.GetString("log.level")
	}
	if viper.IsSet("log.log-dir") {
		config.LogDir = viper.GetString("log.log-dir")
	}
	if viper.IsSet("log.enable-otlp") {
		config.EnableOTLP = viper.GetBool("log.enable-otlp")
	}
	if viper.IsSet("log.otlp-endpoint") {
		config.OTLPEndpoint = viper.GetString("log.otlp-endpoint")
	}
	
	return config
}

// parseDetailedConfig 解析详细配置（向后兼容）
func parseDetailedConfig() *logger.LogsOptions {
	logOptions := logger.DefaultOptions()

	// Configure logging options from viper
	if viper.IsSet("log.type") {
		logOptions.Type = logger.LoggerType(viper.GetString("log.type"))
	}
	if viper.IsSet("log.dynamic") {
		logOptions.Dynamic = viper.GetBool("log.dynamic")
	}
	if viper.IsSet("log.disable-caller") {
		logOptions.DisableCaller = viper.GetBool("log.disable-caller")
	}
	if viper.IsSet("log.disable-stacktrace") {
		logOptions.DisableStacktrace = viper.GetBool("log.disable-stacktrace")
	}
	if viper.IsSet("log.enable-color") {
		logOptions.EnableColor = viper.GetBool("log.enable-color")
	}
	if viper.IsSet("log.level") {
		logOptions.Level = viper.GetString("log.level")
	}
	if viper.IsSet("log.format") {
		logOptions.Format = viper.GetString("log.format")
	}
	if viper.IsSet("log.output-paths") {
		logOptions.OutputPaths = viper.GetStringSlice("log.output-paths")
	}
	if viper.IsSet("log.error-output-paths") {
		logOptions.ErrorOutputPaths = viper.GetStringSlice("log.error-output-paths")
	}
	if viper.IsSet("log.development") {
		logOptions.Development = viper.GetBool("log.development")
	}
	if viper.IsSet("log.encoding") {
		logOptions.Encoding = viper.GetString("log.encoding")
	}
	if viper.IsSet("log.max-size") {
		logOptions.MaxSize = viper.GetInt("log.max-size")
	}
	if viper.IsSet("log.max-age") {
		logOptions.MaxAge = viper.GetInt("log.max-age")
	}
	if viper.IsSet("log.max-backups") {
		logOptions.MaxBackups = viper.GetInt("log.max-backups")
	}
	if viper.IsSet("log.compress") {
		logOptions.Compress = viper.GetBool("log.compress")
	}
	if viper.IsSet("log.caller-skip") {
		logOptions.CallerSkip = viper.GetInt("log.caller-skip")
	}

	// 处理编码器配置
	if viper.IsSet("log.encoder-config") {
		encoderConfig := &logger.EncoderConfig{}
		if viper.IsSet("log.encoder-config.time-key") {
			encoderConfig.TimeKey = viper.GetString("log.encoder-config.time-key")
		}
		if viper.IsSet("log.encoder-config.level-key") {
			encoderConfig.LevelKey = viper.GetString("log.encoder-config.level-key")
		}
		if viper.IsSet("log.encoder-config.message-key") {
			encoderConfig.MessageKey = viper.GetString("log.encoder-config.message-key")
		}
		if viper.IsSet("log.encoder-config.caller-key") {
			encoderConfig.CallerKey = viper.GetString("log.encoder-config.caller-key")
		}
		if viper.IsSet("log.encoder-config.stacktrace-key") {
			encoderConfig.StacktraceKey = viper.GetString("log.encoder-config.stacktrace-key")
		}
		if viper.IsSet("log.encoder-config.function-key") {
			encoderConfig.FunctionKey = viper.GetString("log.encoder-config.function-key")
		}
		if viper.IsSet("log.encoder-config.time-encoder") {
			encoderConfig.TimeEncoder = viper.GetString("log.encoder-config.time-encoder")
		}
		if viper.IsSet("log.encoder-config.level-encoder") {
			encoderConfig.LevelEncoder = viper.GetString("log.encoder-config.level-encoder")
		}
		if viper.IsSet("log.encoder-config.caller-encoder") {
			encoderConfig.CallerEncoder = viper.GetString("log.encoder-config.caller-encoder")
		}
		logOptions.EncoderConfig = encoderConfig
	}

	return logOptions
}

// addVersionInfoToLogger 添加版本信息到日志器
func addVersionInfoToLogger(logOptions *logger.LogsOptions) {
	// 处理初始字段配置
	if logOptions.InitialFields == nil {
		logOptions.InitialFields = make(map[string]interface{})
	}

	versionInfo := version.Get()
	
	// 添加版本信息字段
	if _, exists := logOptions.InitialFields["service"]; !exists {
		logOptions.InitialFields["service"] = versionInfo.ServiceName
	}
	if _, exists := logOptions.InitialFields["version"]; !exists {
		logOptions.InitialFields["version"] = versionInfo.GitVersion
	}
	if _, exists := logOptions.InitialFields["branch"]; !exists {
		logOptions.InitialFields["branch"] = versionInfo.GitBranch
	}
}

func PrintConfig() {
	for _, key := range viper.AllKeys() {
		logger.Debugw(fmt.Sprintf("CFG: %s=%v", key, viper.Get(key)))
	}
}
