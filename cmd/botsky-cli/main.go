package main

import (
	"botsky/pkg/botsky"
	"context"
	"fmt"

//	"github.com/davhofer/indigo/api/bsky"
)

// note: this is just for testing/debugging purposes rn
func main() {
    ctx := context.Background()

    defer fmt.Println("botsky is going to bed...")

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


    botsky.Sleep(1)

    listener := botsky.NewPollingNotificationListener(ctx, client)

    if err := listener.RegisterHandler("replyToMentions", botsky.ExampleHandler); err != nil {
        fmt.Println(err)
        return 
    }

    listener.Start()

    botsky.WaitUntilCancel()


    listener.Stop()

    botsky.Sleep(50)



    /*
    output, err := bsky.NotificationListNotifications(ctx, client.XrpcClient, "", 50, false, []string{}, "")
    fmt.Println(*output.SeenAt)
    for _, notif := range output.Notifications {
        if notif.ReasonSubject == nil {
            continue
        }
        fmt.Println(*notif.ReasonSubject)
        post, _ := client.GetPost(ctx, *notif.ReasonSubject)
        fmt.Println(post)
        fmt.Println()
        fmt.Println()
    }
*/ 

    /*
    _, uri, err := client.NewPost(ctx, "A simple text post", false, "", nil, nil, nil, nil, "", "")
    if err != nil {
        fmt.Println(err)
    }
*/

    /*
    botsky.Sleep(1)
    _, _, err = client.Repost(ctx, "at://did:plc:6gqoupmca6cqjrcjeh7mb3ek/app.bsky.feed.post/3lgvsc277ss23")
    fmt.Println(err)
*/

    /*
    cid, post, err := client.RepoGetPost(ctx, "at://did:plc:a3fiitdzkbaekw34lhfgjzlo/app.bsky.feed.post/3lgxl3lre5k2b")
    if err != nil {
        fmt.Println(err)
        return 
    }

    fmt.Println(cid, post.Text)
*/


/*
    if err := client.RepoDeleteAllPosts(ctx); err != nil {
        fmt.Println("error:", err)
        return
    }
    botsky.Sleep(1)
*/




/*

NOTES:
- some posts don't appear when creating them too quickly? rate limiting? but they show up on PDS, so who is rate limiting? or is it a appview/relay bug?
*/

    
    /*
    _, _, err = client.NewPost(ctx, "is it finall working??", nil, nil, nil, "", false, nil)
    if err != nil {
        fmt.Println("error:", err)
        return
    }
    fmt.Println("post created")
    botsky.Sleep(1)
    */

    /*
    postViews, err := client.GetPostViews(ctx, handle, -1)
    if err != nil {
        fmt.Println("error:", err)
        return
    }

    fmt.Println("got", len(postViews), "posts")



    for _, postView := range postViews {

        var post bsky.FeedPost
        if err := botsky.DecodeRecordAsLexicon(postView.Record, &post); err != nil {
            fmt.Println("error:", err)
            return
        }
        fmt.Println("post:", post.Text)
        fmt.Println("  likes:", *postView.LikeCount)
        fmt.Println("  quotes:", *postView.QuoteCount)
        fmt.Println("  replies:", *postView.ReplyCount)
        fmt.Println("  reposts:", *postView.RepostCount)
    }

    */

}
