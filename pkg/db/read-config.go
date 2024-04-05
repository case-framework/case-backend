package db

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

func ReadDBConfigFromEnv(
	dbLabel string,
	connectionStrEnv string,
	usernameEnv string,
	passwordEnv string,
	connectionPrefixEnv string,
	timeoutEnv string,
	idleConnTimeoutEnv string,
	maxPoolSizeEnv string,
	useNoCursorTimeoutEnv string,
	dbNamePrefixEnv string,
	instanceIDs []string,

) DBConfig {
	connStr := os.Getenv(connectionStrEnv)
	username := os.Getenv(usernameEnv)
	password := os.Getenv(passwordEnv)
	prefix := os.Getenv(connectionPrefixEnv) // Used in test mode
	if connStr == "" || username == "" || password == "" {
		slog.Error("couldn't read DB credentials", slog.String("db", dbLabel))
		panic("couldn't read DB credentials")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv(timeoutEnv))
	if err != nil {
		slog.Error("DB config could not parse timeout", slog.String("error", err.Error()), slog.String(timeoutEnv, os.Getenv(timeoutEnv)), slog.String("db", dbLabel))
		panic(err)
	}

	IdleConnTimeout, err := strconv.Atoi(os.Getenv(idleConnTimeoutEnv))
	if err != nil {
		slog.Error("DB config could not parse idle connection timeout", slog.String("error", err.Error()), slog.String(idleConnTimeoutEnv, os.Getenv(idleConnTimeoutEnv)), slog.String("db", dbLabel))
		panic(err)
	}

	mps, err := strconv.Atoi(os.Getenv(maxPoolSizeEnv))
	MaxPoolSize := uint64(mps)
	if err != nil {
		slog.Error("DB config could not parse max pool size", slog.String("error", err.Error()), slog.String(maxPoolSizeEnv, os.Getenv(maxPoolSizeEnv)), slog.String("db", dbLabel))
		panic(err)
	}

	noCursorTimeout := os.Getenv(useNoCursorTimeoutEnv) == "true"
	DBNamePrefix := os.Getenv(dbNamePrefixEnv)

	return DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		MaxPoolSize:     MaxPoolSize,
		NoCursorTimeout: noCursorTimeout,
		DBNamePrefix:    DBNamePrefix,
		InstanceIDs:     instanceIDs,
	}
}
