package listeners

import (
	"context"
	"fmt"
	"github.com/davhofer/botsky/pkg/botsky"
	"sync"
	"time"
)

// Generic event handler class for the listener.
type Handler[EventT any] func(context.Context, *botsky.Client, []*EventT)

// Generic event listener.
type Listener[EventT any] struct {
	Name            string
	Client          *botsky.Client
	ctx             context.Context
	Active          bool
	Handlers        map[string]Handler[EventT]
	stopSignal      chan bool
	PollingInterval time.Duration
	mutex           sync.Mutex
	pollEventsFunc  func(context.Context, *botsky.Client) ([]*EventT, error) // gets called every PollingInterval seconds to get a list of events which will then be passed to the handlers
}

// Creates a new listener. The pollEvents argument is a function that gets called in order to fetch the newest set of events to be handled.
func NewListener[EventT any](ctx context.Context, client *botsky.Client, name string, pollEvents func(context.Context, *botsky.Client) ([]*EventT, error)) *Listener[EventT] {
	if name == "" {
		name = "Listener"
	}
	return &Listener[EventT]{
		Name:            name,
		Client:          client,
		ctx:             ctx,
		Active:          false,
		Handlers:        make(map[string]Handler[EventT]),
		stopSignal:      make(chan bool, 1),
		PollingInterval: time.Duration(time.Second * 5), // Default polling interval: 5s
		pollEventsFunc:  pollEvents,
	}
}

// Set how frequently the listener polls for new events.
func (l *Listener[EventT]) SetPollingInterval(seconds uint) {
	// set to default
	if seconds == 0 {
		seconds = 5
	}
	restart := false
	if l.Active {
		l.Stop()
		restart = true
	}
	l.PollingInterval = time.Duration(time.Duration(seconds) * time.Second)
	if restart {
		l.Start()
	}
}

// Try to register a new event handler. The id must be unique.
//
// Every registered event handler gets called on the full list of polled events.
func (l *Listener[EventT]) RegisterHandler(id string, handler Handler[EventT]) error {
	if _, exists := l.Handlers[id]; exists {
		return fmt.Errorf("Handler with id %s already exists.", id)
	}
	l.Handlers[id] = handler
	return nil
}

// Deregister (i.e. deactivate) a registered event handler.
func (l *Listener[EventT]) DeregisterHandler(id string) error {
	if _, exists := l.Handlers[id]; !exists {
		return fmt.Errorf("Handler with id %s is not registered.", id)
	}
	delete(l.Handlers, id)
	return nil
}

// Start listening (polling) in the background. This starts a new go routine.
func (l *Listener[EventT]) Start() {
	if l.Active {
		fmt.Println(l.Name, "is already active.")
		return
	}
	l.Active = true
	go l.listen()
}

// Stop listening.
func (l *Listener[EventT]) Stop() {
	if !l.Active {
		fmt.Println(l.Name, "is already stopped.")
		return
	}
	l.stopSignal <- true
	l.Active = false
}

// Continuous loop that listens and distributes polled events to handlers.
// Is run as a goroutine.
func (l *Listener[EventT]) listen() {
	ticker := time.NewTicker(l.PollingInterval)
	fmt.Println(l.Name, "started")
	defer fmt.Println(l.Name, "stopped")
	defer ticker.Stop()

	for {
		select {
		case <-l.stopSignal:
			return
		case <-ticker.C:

			events, err := l.pollEventsFunc(l.ctx, l.Client)
			if err != nil {
				// TODO: logging/error handling...
				fmt.Println(l.Name, "pollAndHandle error:", err)
				continue
			}

			if len(events) == 0 {
				continue
			}

			for id, handler := range l.Handlers {
				// pass in the associated id with the context
				go handler(context.WithValue(l.ctx, "id", id), l.Client, events)
			}

		}
	}
}

/*
handler functions can be closures, to include e.g. pointers to containers for storing results, channels, the client, etc. to handlers

user must take care of errors in handler, e.g. by logging

TODO: cancellation/timeout of handlers?
TODO: logging

should we directly implement specific event handlers? e.g.
OnMention() {}
OnLike() {}
OnReply() {}
etc.
*/
