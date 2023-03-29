package config

type configDefinition struct {
	General    generalDefinition   `toml:"general"`
	Processors processorDefinition `toml:"processors"`
	Worker     workerDefinition    `toml:"worker"`
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
	RouteCalcUrl        string `toml:"routecalc_url"`
	KojiUrl             string `toml:"koji_url"`
	KojiBearerToken     string `toml:"koji_bearer_token"`
}

type processorDefinition struct {
	GolbatEndpoint  string   `toml:"golbat_endpoint"`
	GolbatRawBearer string   `toml:"golbat_raw_bearer"`
	GolbatApiSecret string   `toml:"golbat_api_secret"`
	RawEndpoints    []string `toml:"raw_endpoints"`
}

type workerDefinition struct {
	LoginDelay       int `toml:"login_delay"`
	RoutePartTimeout int `toml:"route_part_timeout"`
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

var Config configDefinition
