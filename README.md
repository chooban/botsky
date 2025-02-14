# Botsky

A Bluesky API client in Go. Use Botsky to easily build advanced API integrations and automated bots.

![GitHub last commit](https://img.shields.io/github/last-commit/davhofer/botsky) ![GitHub Repo stars](https://img.shields.io/github/stars/davhofer/botsky)

## Features

- easily create posts with images, links, mentions, tags etc.
  - automatic detection/parsing of facets (links, mentions, hashtags)
- send and receive chat messages
- notification listeners to react to mentions, replies, etc.
- chat/DM listeners to react to chat messages
- manipulate data on your PDS, read records from other PDSes
- auth management & auto-refresh
- interacting with user profiles and social graph **WIP**

**Note:** This library is under active development, most features are still work in progress.

Complete code and API docs are coming soon.

### Code examples

For more complete examples and details, also check out the [examples here](https://github.com/davhofer/botsky/tree/main/cmd/examples).

For a full demo bot, check out `cmd/examples/helpful-advice-bot`, running live on Bluesky at [@botsky-bot.bsky.social](https://bsky.app/profile/botsky-bot.bsky.social).

Simplified code snippets (without e.g. error handling) for various use cases below.

Initialization and auth:

```go
// Get creds from command line
handle, appkey, err := botsky.GetCLICredentials()
// Or from env variables BOTSKY_HANDLE and BOTSKY_APPKEY
handle, appkey, err = botsky.GetEnvCredentials()
// Set up a client
client, err := botsky.NewClient(ctx, handle, appkey)
err = client.Authenticate(ctx)
```

Creating posts:

```go
// create a post with an image. uri can be a web url or local file
text := "post with an embedded image"
images := []botsky.ImageSource{{Alt: "The github icon", Uri: "https://github.com/fluidicon.png"}}
pb := botsky.NewPostBuilder(text).AddImages(images)
cid, uri, err := client.Post(ctx, pb)
```

```go
// create a post with various (automatically detected) facets, an embedded link, and different post languages
text := "post with #hashtags mentioning @botsky-bot.bsky.social, with an embedded link w/ card, additional tags, and language set to german and english"
pb := botsky.NewPostBuilder(text).
    AddEmbedLink("https://github.com/davhofer/botsky").
    AddTags([]string{"botsky", "is", "happy"}).
    AddLanguage("de").
    AddLanguage("en")
cid, uri, err = client.Post(ctx, pb)
```

```go
// inline links can be both auto detected and added manually
text := "Here are two inline links: https://google.com and a second clickable link"
pb := botsky.NewPostBuilder(text).
    AddInlineLinks([]botsky.InlineLink{{ Text: "a second clickable link", Url: "https://xkcd.com"}}).
cid, uri, err := client.Post(ctx, pb)
```

Create NotificationListener and reply to mentions:

```go
func ExampleMentionHandler(ctx context.Context, client *Client, notifications []*bsky.NotificationListNotifications_Notification) {
    // iterate over all notifications
    for _, notif := range notifications {
        // only consider mentions
        if notif.Reason == "mention" {
            pb := NewPostBuilder("hello :)").ReplyTo(notif.Uri)
            cid, uri, err := client.Post(ctx, pb)
        }
    }
}
func main () {
    // ...
    listener := botsky.NewPollingNotificationListener(ctx, client)
    handlerId := "replyToMentions"
    err := listener.RegisterHandler(handlerId, ExampleMentionHandler)
    listener.Start()
    botsky.WaitUntilCancel()
    listener.Stop()
}
```

Create ChatListener and reply to messages:

```go
func ExampleChatMessageHandler(ctx context.Context, client *botsky.Client, chatElems []*chat.ConvoGetLog_Output_Logs_Elem) {
    // iterate over all new chat log elements
    for _, elem := range chatElems {
        // only consider messages from other people
        if elem.ConvoDefs_LogCreateMessage != nil && elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Sender.Did != client.Did {
            // reply by quoting what they said
            convoId := elem.ConvoDefs_LogCreateMessage.ConvoId
            msgText := elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Text
            reply := "You said: '" + msgText + "'"
            id, rev, err := client.ChatConvoSendMessage(ctx, convoId, reply)
        }
    }
}
func main() {
    // ...
    listener := listeners.NewPollingChatListener(ctx, client)
    err := listener.RegisterHandler("replyToChatMsgs", ExampleChatMessageHandler)
    listener.Start()
    botsky.WaitUntilCancel()
    listener.Stop()
}
```

## Contributing

Issues & pull requests are welcome. For bigger contributions, please open issues to discuss the changes first before submitting a PR. Also, feel free to open issues with feature requests or ideas.

## Why yet another Bluesky API tool/client/bot?

**tldr:** Support for Bluesky automation/bots in Go is not that great yet imo - existing libraries are either purely CLI or support only a small portion of the API and features available. Also, I wanted to do it :)

While there are a bunch of tools and clients out there, my impression when starting this was that the Go ecosystem in particular was quite lacking in terms of tooling and support. I was particularly looking for some sort of Go SDK that allows writing bots or other automated apps, interacting with Bluesky, PDSs, etc. The official Go SDK for atproto, [indigo](https://github.com/bluesky-social/indigo), does not itself provide an API client (it contains an xrpc client, lexicons, and more tho).

There are multiple CLI clients written in Go ([mattn/bsky](https://github.com/mattn/bsky) and [gosky](https://github.com/bluesky-social/indigo/tree/main/cmd/gosky), part of indigo), but those are very much designed for CLI-only use and I wanted a cleaner solution than just wrapping shell commands. Integrating these libraries directly in other Go code would require quite a lot of rewriting I think, so might as well do it from scratch (famous last words).

Lastly, there are [danrusei/gobot-bsky](https://github.com/danrusei/gobot-bsky) and [karalabe/go-bluesky](https://github.com/karalabe/go-bluesky). Both of them generally fit the use case I was looking for, namely being able to cleanly integrate bksy/atproto API automation in Go code. However, while they are great in their own right, they both have quite small coverage of the whole API and seem to be designed for a quite limited range of tasks. For this reason I decided to work on a more feature complete and general library.

And in any case, even if there is some overlap, this project helps me get familiar with atproto and bluesky, so worth it (also: learning go) :)

## A note on federation/decentralization

The library is mainly a Bluesky client, and heavily relies on the API provided by Bluesky (the company) and the Bluesky AppView (except when interacting directly with services not hosted by Bluesky, like alternative PDSes).

## Acknowledgements

This library is partially inspired by and adapted from

- Bluesky Go Bot framework: https://github.com/danrusei/gobot-bsky
- Go-Bluesky client library: https://github.com/karalabe/go-bluesky

## License

The main license for all original work is the MIT license. However, certain parts (adapted from https://github.com/danrusei/gobot-bsky) are published under the Apache 2.0 license.

### TODO/Ideas

- code and api documentation, detailed features overview
- get user profile information, social graph interactions, following & followers, etc.
- demo bot with command/control interface through chat. could set up bot command listeners with authorized users, post creation through chat, etc.
- should we support PDS admin/server/identity functionality? i.e. com.atproto.admin, com.atproto.identity, com.atproto.repo (latter is partially supported)
- further api integration? (lists, feeds, graph, labels, etc.)

- builtin adjustable rate limiting? limits depending on bsky api, pds, ...
- refer to Bluesky guidelines related to API, bots, etc., bots should adhere to guidelines
- reliance on Bluesky's (the company) AppView...
