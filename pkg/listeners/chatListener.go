package listeners

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"

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


const BotDid = "did:plc:a3fiitdzkbaekw34lhfgjzlo"

// example handler that replies to dms by repeating their content
func ExampleChatMessageHandler(ctx context.Context, client *botsky.Client, chatElems []*chat.ConvoGetLog_Output_Logs_Elem) {
	// iterate over all notifications
	for _, elem := range chatElems {
		// only consider messages from other people
        if elem.ConvoDefs_LogCreateMessage != nil && elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Sender.Did != BotDid {
            convoId := elem.ConvoDefs_LogCreateMessage.ConvoId
            msgText := elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Text
            reply := "You said: '" + msgText + "'"
            if _, _, err := client.ChatConvoSendMessage(ctx, convoId, reply); err != nil {
                fmt.Println("Error:", err)
            }
        }
	}
}


// NOTE: ReasonSubject is used by replies and likes, to indicate which of the bots posts it was directed towards
