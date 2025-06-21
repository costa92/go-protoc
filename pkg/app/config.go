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

	"github.com/costa92/go-protoc/v2/pkg/log"
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

	log.Infow("Adding configuration flag", "name", name, "watch", watch, "cfgFile", CfgFile)
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
		log.Infow("Reading configuration file", "name", name, "cfgFile", CfgFile)
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

		log.Debugw("Reading configuration file", "file", CfgFile)

		if err := viper.ReadInConfig(); err != nil {
			log.Errorw(err, "Failed to read configuration file", "file", CfgFile)
		}
		log.Infow("Success to read configuration file", "file", viper.ConfigFileUsed())

		if watch {
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Debugw("Config file changed", "name", e.Name)
			})
		}
	})
}

func PrintConfig() {
	for _, key := range viper.AllKeys() {
		log.Debugw(fmt.Sprintf("CFG: %s=%v", key, viper.Get(key)))
	}
}
