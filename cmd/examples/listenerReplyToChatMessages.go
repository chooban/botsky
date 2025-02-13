package main

import (
	"botsky/pkg/botsky"
	"botsky/pkg/listeners"
	"context"
	"fmt"
)


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

	if err := listener.RegisterHandler("replyToChatMsgs", listeners.ExampleChatMessageHandler); err != nil {
		fmt.Println(err)
		return
	}

	listener.Start()

	botsky.WaitUntilCancel()

	listener.Stop()

	botsky.Sleep(3)
}
