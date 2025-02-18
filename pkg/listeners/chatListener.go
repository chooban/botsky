package listeners

import (
	"context"
	"github.com/davhofer/botsky/pkg/botsky"

	"github.com/davhofer/indigo/api/chat"
)

// Instantiation of the (polling) listenerBase for handling chat logs/events.
type PollingChatListener struct {
	Listener[chat.ConvoGetLog_Output_Logs_Elem]
}

// Returns an set up PollingChatListener.
func NewPollingChatListener(ctx context.Context, client *botsky.Client) *PollingChatListener {
	return &PollingChatListener{*NewListener(ctx, client, "PollingChatListener", pollChatLogs)}
}

// Get all new chat logs since the last check.
func pollChatLogs(ctx context.Context, client *botsky.Client) ([]*chat.ConvoGetLog_Output_Logs_Elem, error) {
	// the "update seen" part happens automatically through updating of the cursor
	return client.ChatGetRecentLogs(ctx)
}
