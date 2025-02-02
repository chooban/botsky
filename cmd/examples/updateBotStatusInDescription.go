package main

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"
)

// This function starts the bot, changes the status in the profile description to active,
// runs it for 30 seconds, and in the end changes the status to inactive
func runBot(handle, appkey, profileDescription string) error {
	ctx := context.Background()
	client, err := botsky.NewClient(ctx, botsky.DefaultServer, handle, appkey)
	if err != nil {
		return err
	}
	err = client.Authenticate(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Authentication successful")

	descriptionActive := profileDescription + "\n\nStatus: Active"
	descriptionInactive := profileDescription + "\n\nStatus: Inactive"

	if err := client.UpdateProfileDescription(ctx, descriptionActive); err != nil {
		return err
	}
	defer client.UpdateProfileDescription(ctx, descriptionInactive)

	botsky.Sleep(30)

	return nil
}

func updateBotStatusInDescription() {

	defer fmt.Println("botsky is going to bed...")

	handle, appkey, err := botsky.GetEnvCredentials()
	if err != nil {
		fmt.Println(err)
		return
	}

	description := `Botsky - Bluesky API client & bot framework written in Go
https://github.com/davhofer/botsky

by @davd.dev`

	if err := runBot(handle, appkey, description); err != nil {
		fmt.Println(err)
	}

}
