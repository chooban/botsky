package main

import (
	"botsky/pkg/botsky"
    "botsky/pkg/listeners"
	"context"
	"fmt"
)

func listenerReplyToMentions() {
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

	botsky.Sleep(1)

	listener := listeners.NewPollingNotificationListener(ctx, client)

	if err := listener.RegisterHandler("replyToMentions", listeners.ExampleMentionHandler); err != nil {
		fmt.Println(err)
		return
	}

	listener.Start()

	botsky.WaitUntilCancel()

	listener.Stop()

	botsky.Sleep(3)

}
