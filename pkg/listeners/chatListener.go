package listeners

import (
	"github.com/davhofer/botsky/pkg/botsky"
	"context"

	"github.com/davhofer/indigo/api/chat"
)

type PollingChatListener struct {
    Listener[chat.ConvoGetLog_Output_Logs_Elem]
}


func pollChatLogs(ctx context.Context, client *botsky.Client) ([]*chat.ConvoGetLog_Output_Logs_Elem, error){
    // the "update seen" part happens automatically through updating of the cursor
    return client.ChatGetRecentLogs(ctx)
}

func NewPollingChatListener(ctx context.Context, client *botsky.Client) *PollingChatListener {
    return &PollingChatListener{*NewListener(ctx, client, "PollingChatListener", pollChatLogs)}
}
