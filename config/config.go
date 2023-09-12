package config

type configDefinition struct {
	General    generalDefinition   `koanf:"general"`
	Processors processorDefinition `koanf:"processors"`
	Worker     workerDefinition    `koanf:"worker"`
	Db         DbDefinition        `koanf:"db"`
	Tuning     tuningDefinition    `koanf:"tuning"`
	Sentry     sentry              `koanf:"sentry"`
	Prometheus prometheus          `koanf:"prometheus"`
	Pyroscope  pyroscope           `koanf:"pyroscope"`
	Koji       koji                `koanf:"koji"`
}

type generalDefinition struct {
	WorkerStatsInterval int    `koanf:"worker_stats_interval"`
	SaveLogs            bool   `koanf:"save_logs"`
	DebugLogging        bool   `koanf:"debug_log"`
	Host                string `koanf:"host"`
	Port                int    `koanf:"port"`
	ApiSecret           string `koanf:"api_secret"`
	BearerToken         string `koanf:"bearer_token"`
	RouteCalcUrl        string `koanf:"routecalc_url"`
}

type processorDefinition struct {
	GolbatEndpoint  string   `koanf:"golbat_endpoint"`
	GolbatRawBearer string   `koanf:"golbat_raw_bearer"`
	GolbatApiSecret string   `koanf:"golbat_api_secret"`
	RawEndpoints    []string `koanf:"raw_endpoints"`
}

type workerDefinition struct {
	LoginDelay       int `koanf:"login_delay"`
	RoutePartTimeout int `koanf:"route_part_timeout"`
}

type DbDefinition struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Name     string `koanf:"name"`
	MaxPool  int    `koanf:"max_pool"`
}

type tuningDefinition struct {
	RecycleGmoLimit          int `koanf:"recycle_gmo_limit"`
	RecycleEncounterLimit    int `koanf:"recycle_encounter_limit"`
	MinimumAccountReuseHours int `koanf:"minimum_account_reuse_hours"`
}

type sentry struct {
	DSN              string  `koanf:"dsn"`
	Debug            bool    `koanf:"debug"`
	SampleRate       float64 `koanf:"sample_rate"`
	EnableTracing    bool    `koanf:"enable_tracing"`
	TracesSampleRate float64 `koanf:"traces_sample_rate"`
}

type pyroscope struct {
	ApplicationName      string `koanf:"application_name"`
	ServerAddress        string `koanf:"server_address"`
	ApiKey               string `koanf:"api_key"`
	Logger               bool   `koanf:"logger"`
	MutexProfileFraction int    `koanf:"mutex_profile_fraction"`
	BlockProfileRate     int    `koanf:"block_profile_rate"`
}

type prometheus struct {
	Enabled    bool      `koanf:"enabled"`
	Token      string    `koanf:"token"`
	BucketSize []float64 `koanf:"bucket_size"`
}

type koji struct {
	Url           string `koanf:"url"`
	BearerToken   string `koanf:"bearer_token"`
	LoadAtStartup bool   `koanf:"load_at_startup"`
	ProjectName   string `koanf:"project_name"`
}

var Config configDefinition
