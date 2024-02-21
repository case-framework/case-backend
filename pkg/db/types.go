package db

type DBConfig struct {
	URI             string
	DBNamePrefix    string
	Timeout         int
	NoCursorTimeout bool
	MaxPoolSize     uint64
	IdleConnTimeout int
	InstanceIDs     []string
}
