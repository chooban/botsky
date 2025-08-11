package main

import (
	"context"
	"fmt"
	"github.com/davhofer/botsky/pkg/botsky"
	"github.com/davhofer/botsky/pkg/listeners"

	"github.com/bluesky-social/indigo/api/chat"
)

// example handler that replies to dms by repeating their content
// gets called by the listener when there are new messages
func ExampleChatMessageHandler(ctx context.Context, client *botsky.Client, chatElems []*chat.ConvoGetLog_Output_Logs_Elem) {
	// iterate over all new chat logs
	for _, elem := range chatElems {
		// only consider messages from other people
		if elem.ConvoDefs_LogCreateMessage != nil && elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Sender.Did != client.Did {
			// reply by quoting what they said
			convoId := elem.ConvoDefs_LogCreateMessage.ConvoId
			msgText := elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Text
			reply := "You said: '" + msgText + "'"
			if _, _, err := client.ChatConvoSendMessage(ctx, convoId, reply); err != nil {
				fmt.Println("Error:", err)
			}
		}
	}
}

// Note: in my testing, seeing the replies pop up in the web interface often took a few seconds/required me to refresh the page
func listenerReplyToChatMessages() {
	ctx := context.Background()

	defer fmt.Println("botsky is going to bed...")

	handle, appkey, err := botsky.GetEnvCredentials()
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := botsky.NewClient(ctx, handle, appkey)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.Authenticate(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Authentication successful")

	listener := listeners.NewPollingChatListener(ctx, client)

	if err := listener.RegisterHandler("replyToChatMsgs", ExampleChatMessageHandler); err != nil {
		fmt.Println(err)
		return
	}

	listener.Start()

	botsky.WaitUntilCancel()

	listener.Stop()

	botsky.Sleep(3)
}
