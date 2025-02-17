package listeners

import (
	"github.com/davhofer/botsky/pkg/botsky"
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

// NOTE for handlers: ReasonSubject is used by replies and likes, to indicate which of the bots posts it was directed towards
