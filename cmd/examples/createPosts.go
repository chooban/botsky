package main

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"
)

func createPosts() {
	defer fmt.Println("botsky is going to bed...")

	ctx := context.Background()

	// authenticate via env variables or cli
	// handle, appkey, err := botsky.GetCLICredentials()
	handle, appkey, err := botsky.GetEnvCredentials()
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := botsky.NewClient(ctx, botsky.DefaultServer, handle, appkey)
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
	botsky.Sleep(2)

	mentionHandle := "@botsky-bot.bsky.social"
	text := fmt.Sprintf("This is a post with #hashtags, mentioning myself %s. It includes an inline-link, as well as an embedded link. Note the additional tags and the post language :)", mentionHandle)
	inlineLinks := []botsky.InlineLink{{Text: "inline-link", Url: "https://xkcd.com"}}

	embeddedLink := "https://github.com/davhofer/botsky"

	// a post including mention, tags, inline link, language, and embedded link
	pb := botsky.NewPostBuilder(text).
		AddInlineLinks(inlineLinks).
		AddEmbedLink(embeddedLink).AddTags([]string{"tagged", "cool"}).AddLanguage("de")

	cid, uri, err := client.Post(ctx, pb)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("posted:", cid, uri)
	}

	botsky.Sleep(2)

	text = "Look at those #beautiful #images :D"

	// image sources include Alt text as well as a uri, which can be either a
	// local file path or a link (links must start with http:// or https://)
	images := []botsky.ImageSource{{Alt: "The github icon", Uri: "https://github.com/fluidicon.png"}}

	// a post wth tags and embedded images
	pb = botsky.NewPostBuilder(text).AddImages(images)
	cid, uri, err = client.Post(ctx, pb)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("posted:", cid, uri)
	}
	botsky.Sleep(2)

	// reply to the previous post
    pb = botsky.NewPostBuilder("this is a reply. also, look at this link: https://google.com").ReplyTo(uri)
	cid, uri, err = client.Post(ctx, pb)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("posted:", cid, uri)
	}
	botsky.Sleep(2)

	// quote the previous post
	pb = botsky.NewPostBuilder("this is a quote of another post").AddQuotedPost(uri)
	cid, uri, err = client.Post(ctx, pb)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("posted:", cid, uri)
	}

	// repost
	cid, uri, err = client.Repost(ctx, "at://did:plc:6gqoupmca6cqjrcjeh7mb3ek/app.bsky.feed.post/3lgvsc277ss23")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("posted:", cid, uri)
	}

}
