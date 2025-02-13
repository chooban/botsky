package botsky

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/davhofer/indigo/api/bsky"
)

type Handler func(context.Context, *Client, []*bsky.NotificationListNotifications_Notification)

type PollingNotificationListener struct {
	Client          *Client
	ctx             context.Context
	Active          bool
	Handlers        map[string]Handler
	stopSignal      chan bool
	PollingInterval time.Duration
	mutex           sync.Mutex
}

func NewPollingNotificationListener(ctx context.Context, client *Client) *PollingNotificationListener {
	return &PollingNotificationListener{
		Client:          client,
		ctx:             ctx,
		Active:          false,
		Handlers:        make(map[string]Handler),
		stopSignal:      make(chan bool, 1),
		PollingInterval: time.Duration(time.Second * 5), // Default polling interval: 5s
	}
}

func (l *PollingNotificationListener) SetPollingInterval(seconds uint) {
	if seconds == 0 {
		seconds = 1
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
func (l *PollingNotificationListener) RegisterHandler(id string, handler Handler) error {
	if _, exists := l.Handlers[id]; exists {
		return fmt.Errorf("Handler with id %s already exists.", id)
	}
	l.Handlers[id] = handler
	return nil
}
func (l *PollingNotificationListener) DeregisterHandler(id string) error {
	if _, exists := l.Handlers[id]; !exists {
		return fmt.Errorf("Handler with id %s is not registered.", id)
	}
	delete(l.Handlers, id)
	return nil
}

// start listening. this starts a new go routine
func (l *PollingNotificationListener) Start() {
	if l.Active {
		fmt.Println("Listener is already active.")
		return
	}
	l.Active = true
	go l.listen()
}

// stop listening
func (l *PollingNotificationListener) Stop() {
	if !l.Active {
		fmt.Println("Listener is already stopped.")
		return
	}
	l.stopSignal <- true
	l.Active = false
}

// continuous loop that listens and distributes notifications to handlers
// is run as a goroutine
func (l *PollingNotificationListener) listen() {
	ticker := time.NewTicker(l.PollingInterval)
	logger.Println("Listener started")
	defer fmt.Println("Listener stopped")
	defer ticker.Stop()

	for {
		select {
		case <-l.stopSignal:
			logger.Println()
			return
		case <-ticker.C:
			// check for new notifications
			count, err := l.Client.NotifGetUnreadCount(l.ctx)
			if err != nil {
				logger.Println("Listener error (NotifGetUnreadCount):", err)
				continue
			}
			if count == 0 {
				continue
			}
			logger.Println("listener:", count, "new notifications")
			// if there are, distribute to handlers
			notifications, err := l.Client.NotifGetNotifications(l.ctx, count)
			if err != nil {
				logger.Println("Listener error (NotifGetNotifications):", err)
				continue
			}

			// mark notifications as seen
			updateSeenInput := bsky.NotificationUpdateSeen_Input{
				SeenAt: time.Now().UTC().Format(time.RFC3339),
			}
			if err := bsky.NotificationUpdateSeen(l.ctx, l.Client.XrpcClient, &updateSeenInput); err != nil {
				logger.Println("Listener error (NotificationUpdateSeen):", err)
				continue
			}

			for id, handler := range l.Handlers {
				// pass in the associated id with the context
				go handler(context.WithValue(l.ctx, "id", id), l.Client, notifications)
			}

		}
	}
}

// example handler that replies to mentions
func ExampleHandler(ctx context.Context, client *Client, notifications []*bsky.NotificationListNotifications_Notification) {
	// iterate over all notifications
	for _, notif := range notifications {
		// only consider mentions
		if notif.Reason == "mention" {
			// Uri (!) is the mentioning post
			pb := NewPostBuilder("hello :)").ReplyTo(notif.Uri)
			cid, uri, err := client.Post(ctx, pb)
			fmt.Println("Posted:", cid, uri, err)
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

// NOTE: ReasonSubject is used by replies and likes, to indicate which of the bots posts it was directed towards
