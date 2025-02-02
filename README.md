# Botsky

A Bluesky API client in Go. Use Botsky to easily build Bluesky API integrations, automated apps, and bots.

---

Provides easy-to-use interfaces for:

- creating posts
- event-/notification listeners to react to mentions, replies, etc.
- manipulating data on your PDS, read records from other PDSes
- interacting with user profiles and social graph **(WIP)**
- interacting with feeds, labelers, and more **(WIP)**

Includes auth management & auto-refresh.

**Note:** This library is under active development, most features are still work in progress.

Feel free to open issues to discuss requests regarding the design and features.

## Why yet another Bluesky API tool/client/bot?

**tldr:** Support for Bluesky automation/bots in Go is not that great yet imo - existing libraries are either purely CLI or support only a small portion of the API and features available. Also, I wanted to do it :)

While there are a bunch of tools and clients out there, my impression when starting this was that the Go ecosystem in particular was quite lacking in terms of tooling and support. I was particularly looking for some sort of Go SDK that allows writing bots or other automated apps, interacting with Bluesky, PDSs, etc. The official Go SDK for atproto, [indigo](https://github.com/bluesky-social/indigo), does not itself provide an API client (it contains an xrpc client, lexicons, and more tho).

There are multiple CLI clients written in Go ([mattn/bsky](https://github.com/mattn/bsky) and [gosky](https://github.com/bluesky-social/indigo/tree/main/cmd/gosky), part of indigo), but those are very much designed for CLI-only use and I wanted a cleaner solution than just wrapping shell commands. Integrating these libraries directly in other Go code would require quite a lot of rewriting I think, so might as well do it from scratch (famous last words).

Lastly, there are [danrusei/gobot-bsky](https://github.com/danrusei/gobot-bsky) and [karalabe/go-bluesky](https://github.com/karalabe/go-bluesky). Both of them generally fit the use case I was looking for, namely being able to cleanly integrate bksy/atproto API automation in Go code. However, while they are great in their own right, they both have quite small coverage of the whole API and seem to be designed for a quite limited range of tasks. For this reason I decided to work on a more feature complete and general library.

And in any case, even if there is some overlap, this project helps me get familiar with atproto and bluesky, so worth it (also: learning go) :)

## Features

Detailed descriptions and API docs coming soon.

### Code examples

For more examples and details, also check out the [examples here](https://github.com/davhofer/botsky/tree/main/cmd/examples).

Initialization and auth:

```go
    // Get creds from command line
    handle, appkey, err := botsky.GetCLICredentials()
    // Or from env variables BOTSKY_HANDLE and BOTSKY_APPKEY
    handle, appkey, err = botsky.GetEnvCredentials()
    // (error handling...)
    // Set up a client interacting with the default server at https://bsky.social
    client, err := botsky.NewClient(ctx, botsky.DefaultServer, handle, appkey)
    // (error handling...)
    err = client.Authenticate(ctx)
```

Creating posts:

```go
text := "post with an embedded image"
images := []botsky.ImageSource{{Alt: "The github icon", Uri: "https://github.com/fluidicon.png"}}
pb := botsky.NewPostBuilder(text).AddImages(images)
cid, uri, err := client.Post(ctx, pb)
```

```go
text := "post with #hashtags mentioning @botsky-bot.bsky.social, an inline-link, an embedded link w/ card, additional tags, and language set to german"
pb := botsky.NewPostBuilder(text).
    AddMentions([]string{"@botsky-bot.bsky.social"}).
    AddInlineLinks([]botsky.InlineLink{{ Text: "inline-link", Url: "https://xkcd.com"}}).
    AddEmbedLink("https://github.com/davhofer/botsky").
    AddTags([]string{"botsky", "is", "happy"}).
    AddLanguage("de")
cid, uri, err = client.Post(ctx, pb)
```

Create NotificationListener and reply to mentions:

```go
listener := botsky.NewPollingNotificationListener(ctx, client)
err := listener.RegisterHandler("replyToMentions", botsky.ExampleHandler)
// (error handling...)
listener.Start()
botsky.WaitUntilCancel()
listener.Stop()
```

The example handler function used to reply to mentions:

```go
func ExampleHandler(ctx context.Context, client *Client, notifications []*bsky.NotificationListNotifications_Notification) {
	// iterate over all notifications
	for _, notif := range notifications {
		// only consider mentions
		if notif.Reason == "mention" {
			// Uri is the mentioning post
			pb := NewPostBuilder("hello :)").ReplyTo(notif.Uri)
			cid, uri, err := client.Post(ctx, pb)
			fmt.Println("Posted:", cid, uri, err)
		}
	}
}
```

## A note on federation/decentralization

The library is mainly a Bluesky client, and heavily relies on the API provided by Bluesky (the company) and the Bluesky AppView (except when interacting directly with services not hosted by Bluesky, like alternative PDSes).

decentralization/federation
=> when possible, interact with atproto instead of Bluesky API

## TODO

- finish PDS/Repo API functionality
- get user profile information, social graph interactions, following & followers, etc.
- further api integration (lists, feeds, graph, labels, etc.)
- code docs, detailed feature overview

- builtin adjustable rate limiting? limits depending on bsky api, pds, ...
- jetStream integration/listener interface?
- refer to Bluesky guidelines related to API, bots, etc., bots should adhere to guidelines
- trust, verification, cryptography: in general the server hosting the PDS is not trusted, should we verify data returned by it?
- reliance on Bluesky's (the company) AppView...

## Acknowledgements

This library is partially inspired by and adapted from

- Bluesky Go Bot framework: https://github.com/danrusei/gobot-bsky
- Go-Bluesky client library: https://github.com/karalabe/go-bluesky
