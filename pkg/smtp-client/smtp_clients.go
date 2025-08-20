package smtp_client

import (
	"crypto/tls"
	"log/slog"
	"net/smtp"
	"strconv"
	"time"

	"github.com/knadh/smtppool"
)

type SmtpClients struct {
	servers        SmtpServerList
	connectionPool []*smtppool.Pool
	counter        uint64
}

func NewSmtpClients(config SmtpServerList) (*SmtpClients, error) {

	sc := &SmtpClients{
		servers:        config,
		counter:        0,
		connectionPool: initConnectionPool(config),
	}
	return sc, nil
}

func initConnectionPool(serverList SmtpServerList) []*smtppool.Pool {
	connectionPools := []*smtppool.Pool{}
	for _, server := range serverList.Servers {
		pool, err := connectToPool(server)
		if err != nil {
			slog.Error("error setting up connection pool", slog.String("error", err.Error()), slog.String("server", server.Address()))

			continue
		} else {
			connectionPools = append(connectionPools, pool)
		}
	}
	if len(connectionPools) < 1 {
		panic("no smtp server connection in the pool")
	}
	return connectionPools
}

func connectToPool(server SmtpServer) (*smtppool.Pool, error) {

	//Set number of concurrent connections here
	auth := smtp.PlainAuth(
		"",
		server.AuthData.Username,
		server.AuthData.Password,
		server.Host,
	)
	if server.AuthData.Username == "" && server.AuthData.Password == "" {
		auth = nil
	}

	tlsOpts := &tls.Config{
		InsecureSkipVerify: server.InsecureSkipVerify,
		ServerName:         server.Host,
	}
	port, err := strconv.Atoi(server.Port)
	if err != nil {
		return nil, err
	}

	pool, err := smtppool.New(smtppool.Opt{
		Host:            server.Host,
		Port:            port,
		MaxConns:        server.Connections,
		IdleTimeout:     time.Duration(server.SendTimeout) * time.Second,
		PoolWaitTimeout: time.Duration(server.SendTimeout) * time.Second,
		TLSConfig:       tlsOpts,
		Auth:            auth,
	})
	// pool, err := email.NewPool(server.Address(), server.Connections, auth, tlsOpts)
	return pool, err
}
