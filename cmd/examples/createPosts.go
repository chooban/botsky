package main

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"
)

func main() {
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
    text := fmt.Sprintf("This is a post with #hashtags, mentioning myself %s. It includes an inline-link, as well as an embedded link :)", mentionHandle)
    mentions := []string{mentionHandle}
    renderHashtags := true
    inlineLinks := []botsky.InlineLink{{ Text: "inline-link", Url: "https://xkcd.com"}}
    // post languages can be set using []{"en", "de", ...}
    // providing nil will default to "en" as post language
    var languages []string = nil

    embeddedLink := "https://github.com/davhofer/botsky"

    // a post including mention, tag, inline link, and embedded link
    cid, uri, err := client.NewPost(ctx, text, renderHashtags, "", mentions, inlineLinks, languages, nil, embeddedLink, "")
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
    cid, uri, err = client.NewPost(ctx, text, renderHashtags, "", nil, nil, nil, images, "", "")
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("posted:", cid, uri)
    }
    botsky.Sleep(2)

    // reply to the previous post
    cid, uri, err = client.NewPost(ctx, "this is a reply", false, uri, nil, nil, nil, nil, "", "")
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("posted:", cid, uri)
    }
    botsky.Sleep(2)
    // quote the previous post
    cid, uri, err = client.NewPost(ctx, "and now I'm quoting a post", false, "", nil, nil, nil, nil, "", uri)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("posted:", cid, uri)
    }
}
