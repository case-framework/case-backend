package db

type DBConfig struct {
	URI              string
	DBNamePrefix     string
	Timeout          int
	NoCursorTimeout  bool
	MaxPoolSize      uint64
	IdleConnTimeout  int
	InstanceIDs      []string
	RunIndexCreation bool
}

type DBConfigYaml struct {
	ConnectionStr      string `yaml:"connection_str"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	ConnectionPrefix   string `yaml:"connection_prefix"`
	Timeout            int    `yaml:"timeout"`
	IdleConnTimeout    int    `yaml:"idle_conn_timeout"`
	MaxPoolSize        int    `yaml:"max_pool_size"`
	UseNoCursorTimeout bool   `yaml:"use_no_cursor_timeout"`
	DBNamePrefix       string `yaml:"db_name_prefix"`
	RunIndexCreation   bool   `yaml:"run_index_creation"`
}
