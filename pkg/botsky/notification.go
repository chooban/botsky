package botsky

import (
	"context"
	"fmt"
    "time"

	"github.com/davhofer/indigo/api/bsky"
)

// TODO: chekc that all functions with cursers that get lists/collections have the abilitiy to iterate and get more

func (c *Client) NotifGetNotifications(ctx context.Context, limit int64) ([]*bsky.NotificationListNotifications_Notification, error) {
	limit = min(100, limit)
	limit = max(1, limit)
	limit = 10
	priority := false
	reasons := []string{}
	output, err := bsky.NotificationListNotifications(ctx, c.xrpcClient, "", limit, priority, reasons, "")
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
	output, err := bsky.NotificationGetUnreadCount(ctx, c.xrpcClient, priority, seenAt)
	if err != nil {
		return 0, fmt.Errorf("Unable to get notification unread count: %v", err)
	}
	return output.Count, nil
}

func (c *Client) NotifUpdateSeenNow(ctx context.Context) error {
    updateSeenInput := bsky.NotificationUpdateSeen_Input{
        SeenAt: time.Now().UTC().Format(time.RFC3339),
    }
    return bsky.NotificationUpdateSeen(ctx, c.xrpcClient, &updateSeenInput)
}
