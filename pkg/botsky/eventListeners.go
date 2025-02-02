package botsky

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/davhofer/indigo/api/bsky"
)

// app.bsky.notification tools


type Handler func(context.Context, *Client, []*bsky.NotificationListNotifications_Notification)


type PollingNotificationListener struct {
    Client *Client
    ctx context.Context
    Active bool
    Handlers map[string]Handler
    stopSignal chan bool 
    PollingInterval time.Duration
    mutex sync.Mutex
}


func NewPollingNotificationListener(ctx context.Context, client *Client) *PollingNotificationListener {
    return &PollingNotificationListener{
        Client: client,
        ctx: ctx,
        Active: false,
        Handlers: make(map[string]Handler),
        stopSignal: make(chan bool, 1),
        PollingInterval: time.Duration(time.Second * 5),
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
    fmt.Println("listener started")
    defer fmt.Println("listener stopping...")
    defer ticker.Stop()

    for {
        select {
        case <- l.stopSignal:
            fmt.Println("listener: stop signal received!")
            return
        case <- ticker.C:
            // check for new notifications 
            fmt.Println("polling...")
            count, err := l.Client.NotifGetUnreadCount(l.ctx) 
            if err != nil {
                // TODO: handle errors, pass to channel?
                fmt.Println("listener error:", err)
                continue
            }
            fmt.Println("listener: new notification count:", count)
            if count == 0 {
                continue
            }
            // if there are, distribute to handlers
            notifications, err := l.Client.NotifGetNotifications(l.ctx, count)
            if err != nil {
                fmt.Println("listener error:", err)
                // TODO: handle errors
                continue 
            }

            // mark notifications as seen
            updateSeenInput := bsky.NotificationUpdateSeen_Input{
                SeenAt: time.Now().UTC().Format(time.RFC3339),
            }
            if err := bsky.NotificationUpdateSeen(l.ctx, l.Client.XrpcClient, &updateSeenInput); err != nil {
                fmt.Println("listener error:", err)
                // TODO: handle errors
                continue 
            }

            fmt.Println("listener: starting handlers")
            for id, handler := range l.Handlers {
                // pass in the associated id with the context
                go handler(context.WithValue(l.ctx, "id", id), l.Client, notifications)
            }


        }
    }
}

// example handler that replies to mentions
func ExampleHandler(ctx context.Context, client *Client, notifications []*bsky.NotificationListNotifications_Notification) {
    defer fmt.Println("example handler done")
    fmt.Println("example handler running")
    // iterate over all notifications
    for _, notif := range notifications {
        fmt.Println(" - - - ")
        fmt.Println("reason:", notif.Reason)
        fmt.Println("isRead:", notif.IsRead)
        fmt.Println("uri:", notif.Uri)
        fmt.Println("reasonSubject:", notif.ReasonSubject)
        fmt.Println("author did:", notif.Author.Did)
        fmt.Println("author handle:", notif.Author.Handle)
        fmt.Println()
        // only consider mentions 
        if notif.Reason == "mention" {
            fmt.Println("handler: mention found!")

            /*
            if notif.ReasonSubject == nil {
                fmt.Println("reason subject nil :(")
                continue 
            }
            */
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


// TODO: chekc that all functions with cursers that get lists/collections have the abilitiy to iterate and get more


// problem is, after we've had a bunch of notifications, filtering by reason will not help anymore.
// because we don't know the reasons of the new notifications until after querying, so if we filter by reason and request n = GetUnreadCount,
// then the filtered out notifications will just be replaced with old, already seen notifications
func (c *Client) NotifGetNotifications(ctx context.Context, limit int64) ([]*bsky.NotificationListNotifications_Notification, error) {
    limit = min(100, limit)
    limit = max(1, limit)
    limit = 10
    priority := false
    reasons := []string{}
    output, err := bsky.NotificationListNotifications(ctx, c.XrpcClient, "", limit, priority, reasons, "")
    if err != nil {
        return nil, fmt.Errorf("Error when calling ListNotifications: %v", err) 
    }

    // TODO: iterate over remaining notifications with cursor
    // (low prio, unlikely that there will be 100+ notifications at a time)

    return output.Notifications, nil
}

func (c *Client) NotifGetUnreadCount(ctx context.Context) (int64, error) {
    priority := false
    seenAt := ""
    output, err := bsky.NotificationGetUnreadCount(ctx, c.XrpcClient, priority, seenAt)
    if err != nil {
        return 0, fmt.Errorf("Unable to get notification unread count: %v", err)
    }
    return output.Count, nil
}
