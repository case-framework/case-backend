package smtp_client

import (
	"errors"
	"log/slog"
	"net/textproto"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"github.com/knadh/smtppool"
)

func (sc *SmtpClients) SendMail(
	to []string,
	subject string,
	htmlContent string,
	overrides *messagingTypes.HeaderOverrides,
) error {
	sc.counter += 1
	if len(sc.connectionPool) < 1 {
		sc.connectionPool = initConnectionPool(sc.servers)
		if len(sc.connectionPool) < 1 {
			return errors.New("no servers defined")
		}
	}

	index := sc.counter % len(sc.connectionPool)
	selectedServer := sc.connectionPool[index]

	From := sc.servers.From
	Sender := sc.servers.Sender
	ReplyTo := sc.servers.ReplyTo

	if overrides != nil {
		if overrides.From != "" {
			From = overrides.From
		}
		if overrides.Sender != "" {
			Sender = overrides.Sender
		}

		if overrides.NoReplyTo {
			ReplyTo = []string{}
		} else if len(overrides.ReplyTo) > 0 {
			ReplyTo = overrides.ReplyTo
		}
	}

	e := smtppool.Email{
		To:      to,
		From:    From,
		Sender:  Sender,
		ReplyTo: ReplyTo,
		Subject: subject,
		HTML:    []byte(htmlContent),
		Headers: textproto.MIMEHeader{},
	}
	err := selectedServer.Send(e)

	if err != nil {
		// close and try to reconnect
		slog.Error("error when trying to send email", slog.String("error", err.Error()))

		pool, errReconnect := connectToPool(sc.servers.Servers[index])
		if errReconnect != nil {
			slog.Error("cannot reconnect pool", slog.String("error", errReconnect.Error()), slog.String("server", sc.servers.Servers[index].Host))
		} else {
			slog.Info("reconnected to pool", slog.String("server", sc.servers.Servers[index].Host))
			sc.connectionPool[index] = pool
		}
	}
	return err
}
