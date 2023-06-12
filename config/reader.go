package config

import (
	"fmt"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"strconv"
	"strings"
)

var k = koanf.New(".")

func ReadConfig() {
	// load default values
	defaultErr := k.Load(structs.Provider(configDefinition{
		General: generalDefinition{
			WorkerStatsInterval: 5,
			SaveLogs:            true,
			Host:                "0.0.0.0",
			Port:                9002,
		},
		Worker: workerDefinition{
			RoutePartTimeout: 150,
			LoginDelay:       20,
		},
		Sentry: sentry{
			SampleRate:       1.0,
			TracesSampleRate: 1.0,
		},
		Pyroscope: pyroscope{
			ApplicationName:      "flygon",
			MutexProfileFraction: 5,
			BlockProfileRate:     5,
		},
		Prometheus: prometheus{
			BucketSize: []float64{.00005, .000075, .0001, .00025, .0005, .00075, .001, .0025, .005, .01, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	}, "koanf"), nil)
	if defaultErr != nil {
		fmt.Println(fmt.Errorf("failed to load default config: %w", defaultErr))
	}

	// read config from file
	readConfigErr := k.Load(file.Provider("config.toml"), toml.Parser())
	if readConfigErr != nil && readConfigErr.Error() != "open config.toml: no such file or directory" {
		fmt.Println(fmt.Errorf("failed to read config file: %w", readConfigErr))
	}

	// read config from env
	envLoadingErr := k.Load(ProviderWithValue("FLYGON.", ".", func(rawKey string, value string, currentMap map[string]interface{}) (string, interface{}) {
		key := strings.ToLower(strings.TrimPrefix(rawKey, "FLYGON."))
		return key, value
	}), nil)
	if envLoadingErr != nil {
		fmt.Println(fmt.Errorf("%w", envLoadingErr))
	}

	unmarshalError := k.Unmarshal("", &Config)
	if unmarshalError != nil {
		panic(fmt.Errorf("failed to Unmarshal config: %w", unmarshalError))
		return
	}
}

func parseEnvVarToSlice(sliceName string, key string, value string, currentMap map[string]interface{}) {
	splitPath := strings.Split(key, ".")
	lastPart := splitPath[len(splitPath)-1]
	index, _ := strconv.Atoi(splitPath[len(splitPath)-2])

	// create the slice if it doesn't exist
	if currentMap[sliceName] == nil {
		currentMap[sliceName] = make([]interface{}, 0)
	}
	// create the element at index
	if len(currentMap[sliceName].([]interface{})) <= index {
		currentMap[sliceName] = append(currentMap[sliceName].([]interface{}), map[string]interface{}{})
	}

	// set the value in map at index in slice
	currentMap[sliceName].([]interface{})[index].(map[string]interface{})[lastPart] = value
}
