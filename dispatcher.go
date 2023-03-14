package outbox

import (
	"log"
	"time"
)

type processor interface {
	ProcessRecords() error
}

type unlocker interface {
	UnlockExpiredMessages() error
}

type cleaner interface {
	RemoveExpiredMessages() error
}

// RetrialPolicy contains the retrial settings
type RetrialPolicy struct {
	MaxSendAttemptsEnabled bool
	MaxSendAttempts        uint
}

// DispatcherSettings defines the set of configurations for the dispatcher
type DispatcherSettings struct {
	ProcessInterval           time.Duration
	LockCheckerInterval       time.Duration
	MaxLockTimeDuration       time.Duration
	CleanupWorkerInterval     time.Duration
	RetrialPolicy             RetrialPolicy
	MessagesRetentionDuration time.Duration
}

// Dispatcher initializes and runs the outbox dispatcher
type Dispatcher struct {
	recordProcessor processor
	recordUnlocker  unlocker
	recordCleaner   cleaner
	settings        DispatcherSettings
}

// NewDispatcher constructor
func NewDispatcher(store Store, broker MessageBroker, settings DispatcherSettings, machineID string) *Dispatcher {
	return &Dispatcher{
		recordProcessor: newProcessor(
			store,
			broker,
			machineID,
			settings.RetrialPolicy,
		),
		recordUnlocker: newRecordUnlocker(
			store,
			settings.MaxLockTimeDuration,
		),
		recordCleaner: newRecordCleaner(
			store,
			settings.MessagesRetentionDuration,
		),
		settings: settings,
	}
}

// Run periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
// and updates the message status accordingly
func (d Dispatcher) Run(errChan chan<- error, doneChan <-chan struct{}) {
	doneProc := make(chan struct{}, 1)
	doneUnlock := make(chan struct{}, 1)
	doneClear := make(chan struct{}, 1)

	go func() {
		<-doneChan
		doneProc <- struct{}{}
		doneUnlock <- struct{}{}
		doneClear <- struct{}{}
	}()

	go d.runRecordProcessor(errChan, doneProc)
	go d.runRecordUnlocker(errChan, doneUnlock)
	go d.runRecordCleaner(errChan, doneClear)
}

// runRecordProcessor processes the unsent records of the store
func (d Dispatcher) runRecordProcessor(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.ProcessInterval)
	for {
		log.Print("Record processor Running")
		err := d.recordProcessor.ProcessRecords()
		if err != nil {
			errChan <- err
		}
		log.Print("Record Processing Finished")

		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record processor")
			return
		}
	}
}

func (d Dispatcher) runRecordUnlocker(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.LockCheckerInterval)
	for {
		log.Print("Record unlocker Running")
		err := d.recordUnlocker.UnlockExpiredMessages()
		if err != nil {
			errChan <- err
		}
		log.Print("Record unlocker Finished")
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record unlocker")
			return

		}
	}
}

func (d Dispatcher) runRecordCleaner(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.CleanupWorkerInterval)
	for {
		log.Print("Record retention cleaner Running")
		err := d.recordCleaner.RemoveExpiredMessages()
		if err != nil {
			errChan <- err
		}
		log.Print("Record retention cleaner Finished")
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record retention cleaner")
			return

		}
	}
}
