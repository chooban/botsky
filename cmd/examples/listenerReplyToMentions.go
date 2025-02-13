package main

import (
	"botsky/pkg/botsky"
	"botsky/pkg/listeners"
	"context"
	"fmt"

	"github.com/davhofer/indigo/api/bsky"
)

// example handler that replies to mentions
// gets called by the listener
func ExampleMentionHandler(ctx context.Context, client *botsky.Client, notifications []*bsky.NotificationListNotifications_Notification) {
	// iterate over all notifications
	for _, notif := range notifications {
		// only consider mentions
		if notif.Reason == "mention" {
			// Uri is the mentioning post
			pb := botsky.NewPostBuilder("hello :)").ReplyTo(notif.Uri)
			cid, uri, err := client.Post(ctx, pb)
			fmt.Println("Posted:", cid, uri, err)
		}
	}
}

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

	if err := listener.RegisterHandler("replyToMentions", ExampleMentionHandler); err != nil {
		fmt.Println(err)
		return
	}

	listener.Start()

	botsky.WaitUntilCancel()

	listener.Stop()

	botsky.Sleep(3)

}
