package config

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mgdi/kongroo-c2/c2/types"

	"github.com/spf13/viper"
)

var Configs map[string]string

func init() {
	Configs = make(map[string]string)
}

var configPath = []string{
	fmt.Sprintf("/etc/%s", types.APPNAME),
	fmt.Sprintf("$HOME/.%s", types.APPNAME),
	".",
}

func InitializeConfigs() {
	log.Debug("Setup configuration")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix(types.CONF_PREFIX)
	viper.AutomaticEnv()

	for _, path := range configPath {
		viper.AddConfigPath(path)
	}

	viper.ReadInConfig()

	log.Debugf("Searching configuration file in: %s", configPath)

	for _, option := range viper.AllKeys() {
		configOption := fmt.Sprintf("%s_%s", types.CONF_PREFIX, strings.ToUpper(strings.Replace(option, ".", "_", -1)))
		log.WithFields(log.Fields{
			configOption: option,
		}).Debugf("Converting ENV variables to program variable")
		viper.BindEnv(option, configOption)

		if !viper.IsSet(option) {
			log.Fatalf("Config option: %s not found.", option)
		}
		Configs[option] = getConfig(option)
	}
}

func getConfig(config string) (value string) {
	value = viper.GetString(config)

	log.Debugf("Getting configuration %s...", config)

	log.WithFields(log.Fields{
		"SECTION": "CONFIGURATION",
		config:    value,
	}).Debug("Reading configuration value...")

	return
}
