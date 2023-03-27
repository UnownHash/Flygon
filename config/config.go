package config

type configDefinition struct {
	General    generalDefinition   `toml:"general"`
	Processors processorDefinition `toml:"processors"`
	Db         DbDefinition        `toml:"db"`
	Sentry     sentry              `toml:"sentry"`
	Pyroscope  pyroscope           `toml:"pyroscope"`
}

type generalDefinition struct {
	WorkerStatsInterval int    `toml:"worker_stats_interval"`
	SaveLogs            bool   `toml:"save_logs"`
	DebugLogging        bool   `toml:"debug_log"`
	Host                string `toml:"host"`
	Port                int    `toml:"port"`
	ApiSecret           string `tom:"api_secret"`
	BearerToken         string `tom:"bearer_token"`
	LoginDelay          int    `toml:"login_delay"`
	RouteCalcUrl        string `toml:"routecalc_url"`
	KojiUrl             string `toml:"koji_url"`
	KojiBearerToken     string `toml:"koji_bearer_token"`
	DisableFortLookup   bool   `toml:"disable_fort_lookup"`
}

type processorDefinition struct {
	GolbatEndpoint  string   `toml:"golbat_endpoint"`
	GolbatRawBearer string   `toml:"golbat_raw_bearer"`
	GolbatApiSecret string   `toml:"golbat_api_secret"`
	RawEndpoints    []string `toml:"raw_endpoints"`
}

type DbDefinition struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	MaxPool  int `toml:"max_pool"`
}

type sentry struct {
	DSN              string  `toml:"dsn"`
	Debug            bool    `toml:"debug"`
	SampleRate       float64 `toml:"sample_rate"`
	EnableTracing    bool    `toml:"enable_tracing"`
	TracesSampleRate float64 `toml:"traces_sample_rate"`
}

type pyroscope struct {
	ApplicationName      string `toml:"application_name"`
	ServerAddress        string `toml:"server_address"`
	ApiKey               string `toml:"api_key"`
	Logger               bool   `toml:"logger"`
	MutexProfileFraction int    `toml:"mutex_profile_fraction"`
	BlockProfileRate     int    `toml:"block_profile_rate"`
}

var Config = configDefinition{
	General: generalDefinition{
		WorkerStatsInterval: 5,
		SaveLogs:            true,
		Host:                "0.0.0.0",
		Port:                9002,
		LoginDelay:          20,
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
}
