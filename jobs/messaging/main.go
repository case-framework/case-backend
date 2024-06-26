package main

import (
	"log/slog"
	"sync"
	"time"
)

const (
	OUTGOING_EMAILS_BATCH_SIZE = 10

	MAX_FAILED_ATTEMPTS_BEFORE_STOP = 100
)

func main() {
	slog.Info("Starting messaging job")
	start := time.Now()

	var wg sync.WaitGroup

	if conf.RunTasks.ProcessOutgoingEmails {
		wg.Add(1)
		go handleOutgoingMessages(&wg)
	}

	if conf.RunTasks.ScheduleHandler {
		wg.Add(1)
		go handleScheduledMessages(&wg)
	}

	if conf.RunTasks.StudyMessagesHandler {
		wg.Add(1)
		go handleParticipantMessages(&wg)
	}

	if conf.RunTasks.ResearcherMessagesHandler {
		wg.Add(1)
		go handleResearcherNotifications(&wg)
	}

	wg.Wait()
	slog.Info("Messaging job completed", slog.String("duration", time.Since(start).String()))
}
