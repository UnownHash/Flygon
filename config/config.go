package config

type configDefinition struct {
	General    generalDefinition   `mapstructure:"general"`
	Processors processorDefinition `mapstructure:"processors"`
	Worker     workerDefinition    `mapstructure:"worker"`
	Db         DbDefinition        `mapstructure:"db"`
	Sentry     sentry              `mapstructure:"sentry"`
	Prometheus prometheus          `mapstructure:"prometheus"`
	Pyroscope  pyroscope           `mapstructure:"pyroscope"`
	Koji       koji                `mapstructure:"koji"`
}

type generalDefinition struct {
	WorkerStatsInterval int    `mapstructure:"worker_stats_interval"`
	SaveLogs            bool   `mapstructure:"save_logs"`
	DebugLogging        bool   `mapstructure:"debug_log"`
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	ApiSecret           string `mapstructure:"api_secret"`
	BearerToken         string `mapstructure:"bearer_token"`
	RouteCalcUrl        string `mapstructure:"routecalc_url"`
}

type processorDefinition struct {
	GolbatEndpoint  string   `mapstructure:"golbat_endpoint"`
	GolbatRawBearer string   `mapstructure:"golbat_raw_bearer"`
	GolbatApiSecret string   `mapstructure:"golbat_api_secret"`
	RawEndpoints    []string `mapstructure:"raw_endpoints"`
}

type workerDefinition struct {
	LoginDelay       int `mapstructure:"login_delay"`
	RoutePartTimeout int `mapstructure:"route_part_timeout"`
}

type DbDefinition struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	MaxPool  int    `mapstructure:"max_pool"`
}

type sentry struct {
	DSN              string  `mapstructure:"dsn"`
	Debug            bool    `mapstructure:"debug"`
	SampleRate       float64 `mapstructure:"sample_rate"`
	EnableTracing    bool    `mapstructure:"enable_tracing"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
}

type pyroscope struct {
	ApplicationName      string `mapstructure:"application_name"`
	ServerAddress        string `mapstructure:"server_address"`
	ApiKey               string `mapstructure:"api_key"`
	Logger               bool   `mapstructure:"logger"`
	MutexProfileFraction int    `mapstructure:"mutex_profile_fraction"`
	BlockProfileRate     int    `mapstructure:"block_profile_rate"`
}

type prometheus struct {
	Enabled    bool      `mapstructure:"enabled"`
	Token      string    `mapstructure:"token"`
	BucketSize []float64 `mapstructure:"bucket_size"`
}

type koji struct {
	Url           string `mapstructure:"url"`
	BearerToken   string `mapstructure:"bearer_token"`
	LoadAtStartup bool   `mapstructure:"load_at_startup"`
	ProjectName   string `mapstructure:"project_name"`
}

var Config = configDefinition{
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
}
