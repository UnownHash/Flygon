package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

func ReadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	// configure how to read environment variables
	viper.SetEnvPrefix("flygon")
	viper.AutomaticEnv()
	stringReplacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(stringReplacer)

	setDefaults()
	readConfigErr := viper.ReadInConfig()
	if readConfigErr != nil {
		panic(fmt.Errorf("failed to read config file: %w", readConfigErr))
	}
	unmarshalErr := viper.Unmarshal(&Config)
	if unmarshalErr != nil {
		panic(fmt.Errorf("failed to parse config file: %w", unmarshalErr))
	}
}

func setDefaults() {
	viper.SetDefault("general.worker_stats_interval", 5)
	viper.SetDefault("general.save_logs", true)
	viper.SetDefault("general.host", "0.0.0.0")
	viper.SetDefault("general.port", 9002)

	viper.SetDefault("worker.route_part_timeout", 150)
	viper.SetDefault("worker.login_delay", 20)
	viper.SetDefault("sentry.sample_rate", 1.0)
	viper.SetDefault("sentry.traces_sample_rate", 1.0)
	viper.SetDefault("pyroscope.application_name", "flygon")
	viper.SetDefault("pyroscope.mutex_profile_fraction", 5)
	viper.SetDefault("pyroscope.block_profile_rate", 5)
	viper.SetDefault("prometheus.bucket_size", []float64{.00005, .000075, .0001, .00025, .0005, .00075, .001, .0025, .005, .01, .05, .1, .25, .5, 1, 2.5, 5, 10})
}
