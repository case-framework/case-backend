package main

import "sync"

func handleParticipantMessages(wg *sync.WaitGroup) {
	defer wg.Done()

}
