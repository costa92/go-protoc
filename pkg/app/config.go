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

	logger.Infow("Adding configuration flag", "name", name, "watch", watch, "cfgFile", CfgFile)
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

	// 处理初始字段配置 - 添加到所有日志条目的字段
	if viper.IsSet("log.initial-fields") {
		logOptions.InitialFields = make(map[string]interface{})
		initialFieldsMap := viper.GetStringMap("log.initial-fields")
		for k, v := range initialFieldsMap {
			logOptions.InitialFields[k] = v
		}
	}

	// 确保始终有服务信息，如果配置文件中没有设置则使用版本包的默认值
	if logOptions.InitialFields == nil {
		logOptions.InitialFields = make(map[string]interface{})
	}

	versionInfo := version.Get()
	
	// 如果配置中没有 service 字段，使用版本包的默认服务名
	if _, exists := logOptions.InitialFields["service"]; !exists {
		logOptions.InitialFields["service"] = versionInfo.ServiceName
	}
	
	// 如果配置中没有 version 字段，使用版本包的版本信息
	if _, exists := logOptions.InitialFields["version"]; !exists {
		logOptions.InitialFields["version"] = versionInfo.GitVersion
	}
	
	// 如果配置中没有 branch 字段，使用版本包的分支信息
	if _, exists := logOptions.InitialFields["branch"]; !exists {
		logOptions.InitialFields["branch"] = versionInfo.GitBranch
	}

	// Initialize the global logger
	if globalLogger, err := logger.NewLogger(logOptions); err != nil {
		logger.Errorw("Failed to reinitialize logger from config", "error", err)
	} else {
		logger.SetDefaultLogger(globalLogger)
		logger.Infow("Logger reinitialized from configuration file", 
			"level", logOptions.Level, 
			"format", logOptions.Format,
			"service", versionInfo.ServiceName)
	}
}

func PrintConfig() {
	for _, key := range viper.AllKeys() {
		logger.Debugw(fmt.Sprintf("CFG: %s=%v", key, viper.Get(key)))
	}
}
