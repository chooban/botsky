package listeners

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"

	"github.com/davhofer/indigo/api/bsky"
)

type PollingNotificationListener struct {
    Listener[bsky.NotificationListNotifications_Notification]
}


func pollNotifications(ctx context.Context, client *botsky.Client) ([]*bsky.NotificationListNotifications_Notification, error){
    count, err := client.NotifGetUnreadCount(ctx)
    if err != nil {
        return nil, err
    }
    if count == 0 {
        return []*bsky.NotificationListNotifications_Notification{}, nil
    }

    fmt.Println("listener:", count, "new notifications")

    notifications, err := client.NotifGetNotifications(ctx, count)
    if err != nil {
        return nil, err
    }

    // mark notifications as seen
    if err := client.NotifUpdateSeenNow(ctx); err != nil {
        return nil, err
    }

    return notifications, nil
}

func NewPollingNotificationListener(ctx context.Context, client *botsky.Client) *PollingNotificationListener {
    return &PollingNotificationListener{*NewListener(ctx, client, "PollingNotificationListener", pollNotifications)}
}

// example handler that replies to mentions
func ExampleMentionHandler(ctx context.Context, client *botsky.Client, notifications []*bsky.NotificationListNotifications_Notification) {
	// iterate over all notifications
	for _, notif := range notifications {
		// only consider mentions
		if notif.Reason == "mention" {
			// Uri (!) is the mentioning post
			pb := botsky.NewPostBuilder("hello :)").ReplyTo(notif.Uri)
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
