package main

import (
	"botsky/pkg/botsky"
	"botsky/pkg/listeners"
	"fmt"
	"strings"

	"context"
	"encoding/json"
	"math/rand/v2"
	"net/http"

	"github.com/davhofer/indigo/api/bsky"
	"github.com/davhofer/indigo/api/chat"
)

type Slip struct {
    Id int `json:"id"`
    Advice string `json:"advice"`
}
type Response struct {
    Slip Slip `json:"slip"`
}

func MentionHandler(ctx context.Context, client *botsky.Client, notifications []*bsky.NotificationListNotifications_Notification) {

	// iterate over all notifications
	for _, notif := range notifications {
		// only consider mentions
		if notif.Reason == "mention" {
            fmt.Println("mention received")
			// Uri is the mentioning post
            post, err := client.GetPost(ctx, notif.Uri)
            if err != nil {
                fmt.Println(err)
            }

            textLower := strings.ToLower(post.Text)
            if (strings.Contains(textLower, "advice") || strings.Contains(textLower, "help")) {
			    pb := botsky.NewPostBuilder("gotcha, sliding into those DMs").ReplyTo(notif.Uri)
			    _, _, err := client.Post(ctx, pb)
                if err != nil {
                    fmt.Println(err)
                    return
                }

                // slide into DMs
                authorDid := notif.Author.Did

                if _,_, err := client.ChatSendMessage(ctx, authorDid, "you ready for some great advice?"); err != nil {
                    fmt.Println("chat error", err)
                    fmt.Println(err.Error())

                    if strings.Contains(err.Error(), "recipient requires incoming messages to come from someone they follow") {
                        pb := botsky.NewPostBuilder("you gotta let me message you, either follow me or open up DMs in your chat settings, then try again").ReplyTo(notif.Uri)
                        client.Post(ctx, pb)

                    } else if strings.Contains(err.Error(), "recipient has disabled incoming messages") {
                        pb := botsky.NewPostBuilder("you gotta let me message you, change your chat settings and maybe follow me, then try again").ReplyTo(notif.Uri)
                        client.Post(ctx, pb)
                    }
                    return
                }

                advice, err := getAdvice()
                if err != nil {
                    fmt.Println(err)
                    return
                }
                _,_, err = client.ChatSendMessage(ctx, authorDid, "As my mama used to say, " + strings.ToLower(advice))  
                if err != nil {
                    fmt.Println(err)
                    return
                }
                client.ChatSendMessage(ctx, authorDid, "you're welcome")  
                client.ChatSendMessage(ctx, authorDid, "alright gotta go, the world needs me")  

            } else {
			    pb := botsky.NewPostBuilder("idk what you want from me...\nlet me know if you need some great advice").ReplyTo(notif.Uri)
			    _, _, err := client.Post(ctx, pb)
                if err != nil {
                    fmt.Println(err)
                }
            }
		}
	}
}

func ChatMessageHandler(ctx context.Context, client *botsky.Client, chatElems []*chat.ConvoGetLog_Output_Logs_Elem) {
	for _, elem := range chatElems {
        if elem.ConvoDefs_LogCreateMessage != nil && elem.ConvoDefs_LogCreateMessage.Message.ConvoDefs_MessageView.Sender.Did != client.Did {
            convoId := elem.ConvoDefs_LogCreateMessage.ConvoId
            reply := "sorry I'm way too busy for you right now"
            if _, _, err := client.ChatConvoSendMessage(ctx, convoId, reply); err != nil {
                fmt.Println("Error:", err)
                continue
            }
        }
	}
}

func getAdvice() (string, error) {
    // TODO: just download all of them
    // 1 - 224
    id := rand.IntN(224)+1
    url := fmt.Sprintf("https://api.adviceslip.com/advice/%d", id)
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("Error: HTTP Status code != 200: %s", resp.Status)
    }

    var r Response
    if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
        return "", err
    }
    fmt.Println("advice:", r.Slip.Advice)
    return r.Slip.Advice, nil
}
// TBD
func main() {

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

	mentionListener := listeners.NewPollingNotificationListener(ctx, client)

	if err := mentionListener.RegisterHandler("replyToMentions", MentionHandler); err != nil {
		fmt.Println(err)
		return
	}
	chatListener := listeners.NewPollingChatListener(ctx, client)

	if err := chatListener.RegisterHandler("replyToChatMsgs", ChatMessageHandler); err != nil {
		fmt.Println(err)
		return
	}

	mentionListener.Start()
    chatListener.Start()

	botsky.WaitUntilCancel()

	mentionListener.Stop()

	chatListener.Stop()
	botsky.Sleep(3)
}
