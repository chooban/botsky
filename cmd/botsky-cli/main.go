package main

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"
	"time"

)

// note: this is just for testing/debugging purposes rn
func main() {
    ctx := context.Background()

    defer fmt.Println("botsky is going to bed...")

    handle, appkey := botsky.GetEnvCredentials()

    if handle == "" || appkey == "" {
        fmt.Println("Handle and AppKey for authentication need to be accessible as environment variables BOTSKY_HANDLE and BOTSKY_APPKEY")
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

    time.Sleep(1 * time.Second)

/*

NOTES:
- some posts don't appear when creating them too quickly? rate limiting? but they show up on PDS, so who is rate limiting? or is it a appview/relay bug?

*/

    fmt.Println()

    time.Sleep(3 * time.Second)

    fmt.Println("Testing post generation...")

    var cid , url string

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "This is a simple text post", nil, nil, nil, "", false, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    cid, url, err = client.NewPost(ctx, "This is a #post with #hashtags.\nCool#stuff #hmm", nil, nil, nil, "", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    cid, url, err = client.NewPost(ctx, "Hashtags but dont parse. This is a #post with #hashtags.\nCool#stuff #hmm", nil, nil, nil, "", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)
    cid, url, err = client.NewPost(ctx, "mentioning @davd.dev", []string{"@davd.dev"}, nil, nil, "", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)

    cid, url, err = client.NewPost(ctx, "I'll mention myself @botsky-bot.bsky.social ", []string{"@botsky-bot.bsky.social"}, nil, nil, "", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "No mention here @botsky-bot.bsky.social ", nil, nil, nil, "", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "Inline urls anyone?", nil, []botsky.InlineLink{{ Text: "Inline", Url: "https://brave.com"}, { Text: "anyone", Url: "https://google.com"}}, nil, "", false, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "Images...", nil, nil, []botsky.ImageSource{{ Alt: "Local image file.", Uri: "/home/david/Pictures/Screenshots/screenshot.png"}, { Alt: "This is the alt text. Github.", Uri: "https://github.com/fluidicon.png"}}, "", false, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "Embedded url ", nil, nil, nil, "https://github.com/davhofer/botsky", false, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "Everything #hashtag with #image @botsky-bot.bsky.social @davd.dev", []string{"@botsky-bot.bsky.social", "@davd.dev"}, []botsky.InlineLink{{ Text: "Everything", Url: "https://brave.com"}, { Text: "Everything", Url: "https://google.com"}}, []botsky.ImageSource{{ Alt: "Local image file.", Uri: "/home/david/Pictures/Screenshots/screenshot.png"}, { Alt: "This is the alt text. Github.", Uri: "https://github.com/fluidicon.png"}}, "", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

    time.Sleep(1 * time.Second)


    cid, url, err = client.NewPost(ctx, "Everything #hashtag with #embedlink @botsky-bot.bsky.social @davd.dev", []string{"@botsky-bot.bsky.social", "@davd.dev"}, []botsky.InlineLink{{ Text: "Everything", Url: "https://brave.com"}, { Text: "Everything", Url: "https://google.com"}}, nil, "https://github.com/davhofer/botsky", true, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("posted", cid, url)

}
