package config

type configDefinition struct {
	General    generalDefinition
	Processors processorDefinition
	Db         DbDefinition `toml:"db"`
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
	PoolSize int `toml:"pool_size"`
}

var Config = configDefinition{
	General: generalDefinition{
		WorkerStatsInterval: 5,
		SaveLogs:            true,
		Host:                "0.0.0.0",
		Port:                9001,
		LoginDelay:          20,
	},
}
