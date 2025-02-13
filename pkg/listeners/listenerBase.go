package listeners

import (
    "botsky/pkg/botsky"
	"context"
	"fmt"
	"sync"
	"time"
)

type Handler[EventT any] func(context.Context, *botsky.Client, []*EventT)

type Listener[EventT any] struct {
    Name            string
	Client          *botsky.Client
	ctx             context.Context
	Active          bool
	Handlers        map[string]Handler[EventT]
	stopSignal      chan bool
	PollingInterval time.Duration
	mutex           sync.Mutex
    pollEventsFunc func(context.Context, *botsky.Client) ([]*EventT, error) // gets called every PollingInterval seconds to get a list of events which will then be passed to the handlers 
}

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

// try to register handler, make sure id is unique
func (l *Listener[EventT]) RegisterHandler(id string, handler Handler[EventT]) error {
	if _, exists := l.Handlers[id]; exists {
		return fmt.Errorf("Handler with id %s already exists.", id)
	}
	l.Handlers[id] = handler
	return nil
}
func (l *Listener[EventT]) DeregisterHandler(id string) error {
	if _, exists := l.Handlers[id]; !exists {
		return fmt.Errorf("Handler with id %s is not registered.", id)
	}
	delete(l.Handlers, id)
	return nil
}

// start listening in the background. this starts a new go routine
func (l *Listener[EventT]) Start() {
	if l.Active {
		fmt.Println(l.Name, "is already active.")
		return
	}
	l.Active = true
	go l.listen()
}

// stop listening
func (l *Listener[EventT]) Stop() {
	if !l.Active {
		fmt.Println(l.Name, "is already stopped.")
		return
	}
	l.stopSignal <- true
	l.Active = false
}

// continuous loop that listens and distributes notifications to handlers
// is run as a goroutine
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
            
			for id, handler := range l.Handlers {
				// pass in the associated id with the context
				go handler(context.WithValue(l.ctx, "id", id), l.Client, events)
			}

		}
	}
}


/*
handler functions can be closures, to include e.g. pointers to containers for storing results, channels, the client, etc. to handlers

TODO: error handling? let user handle channels?
TODO: pass the client directly to handler, or let them include it with closure? not every handler will need client, but many...
TODO: cancellation/timeout of handlers

should we directly implement specific event handlers? e.g.
OnMention() {}
OnLike() {}
OnReply() {}
etc.
*/

