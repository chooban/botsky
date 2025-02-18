package botsky

import (
	"context"
	"fmt"

	"github.com/davhofer/indigo/api/chat"
)

// Update for the given account whether it can initiate DMs or not.
func (c *Client) ChatUpdateActorAccess(ctx context.Context, handleOrDid string, allowAccess bool) error {
	did, err := c.ResolveHandle(ctx, handleOrDid)
	if err != nil {
		return err
	}
	return chat.ModerationUpdateActorAccess(ctx, c.chatClient, &chat.ModerationUpdateActorAccess_Input{
		Actor:       did,
		AllowAccess: allowAccess,
	})
}

// Get number of unread messages in the given conversation.
func (c *Client) ChatConvoGetUnreadMessageCount(ctx context.Context, convoId string) (int64, error) {
	convo, err := c.ChatGetConvo(ctx, convoId)
	if err != nil {
		return 0, err
	}
	return convo.UnreadCount, nil
}

// Set the message to "read" in the given conversation. Set pointer to nil in order to set the status of all messages in the conversation.
func (c *Client) ChatConvoUpdateRead(ctx context.Context, convoId string, messageId *string) error {
	// messageId is an optional pointer, if only a single message should be updated to read
	_, err := chat.ConvoUpdateRead(ctx, c.chatClient, &chat.ConvoUpdateRead_Input{
		ConvoId:   convoId,
		MessageId: messageId,
	})
	return err
}

// Get the conversation including exactly the provided accounts, or create a new one if it doesn't exist.
func (c *Client) ChatGetConvoForMembers(ctx context.Context, handlesOrDids []string) (*chat.ConvoDefs_ConvoView, error) {
	var dids []string
	for _, handleOrDid := range handlesOrDids {
		did, err := c.ResolveHandle(ctx, handleOrDid)
		if err != nil {
			return nil, fmt.Errorf("ChatGetConvoForMembers error: %v", err)
		}
		dids = append(dids, did)
	}

	// TODO: does this require a handle?
	convoOutput, err := chat.ConvoGetConvoForMembers(ctx, c.chatClient, dids)
	if err != nil {
		return nil, fmt.Errorf("ChatGetConvoForMembers error: %v", err)
	}
	return convoOutput.Convo, nil
}

// Get the conversation by id.
func (c *Client) ChatGetConvo(ctx context.Context, convoId string) (*chat.ConvoDefs_ConvoView, error) {

	convoOutput, err := chat.ConvoGetConvo(ctx, c.chatClient, convoId)
	if err != nil {
		return nil, fmt.Errorf("ChatGetConvo error: %v", err)
	}
	return convoOutput.Convo, nil
}

// Send a text message to the given conversation.
func (c *Client) ChatConvoSendMessage(ctx context.Context, convoId string, message string) (string, string, error) {
	input := chat.ConvoSendMessage_Input{
		ConvoId: convoId,
		Message: &chat.ConvoDefs_MessageInput{
			Text: message,
		},
	}
	msgView, err := chat.ConvoSendMessage(ctx, c.chatClient, &input)
	if err != nil {
		return "", "", fmt.Errorf("ChatSendMessage error: %v", err)
	}
	return msgView.Id, msgView.Rev, nil
}

// List all conversations.
func (c *Client) ChatListConvos(ctx context.Context) ([]*chat.ConvoDefs_ConvoView, error) {
	var convos []*chat.ConvoDefs_ConvoView
	cursor, lastId := "", ""

	// iterate until we got all convos
	for {
		// query repo for collection with updated cursor
		output, err := chat.ConvoListConvos(ctx, c.chatClient, cursor, 100)
		if err != nil {
			return nil, fmt.Errorf("ChatListConvos error: %v", err)
		}

		// stop if no records returned
		// or we get one we've already seen (maybe a repo doesn't support cursor?)
		if len(output.Convos) == 0 || lastId == output.Convos[len(output.Convos)-1].Id {
			break
		}
		lastId = output.Convos[len(output.Convos)-1].Id

		// store all record uris
		convos = append(convos, output.Convos...)

		// update cursor
		cursor = *output.Cursor
	}

	return convos, nil
}

// Get all messages in the given conversation.
func (c *Client) ChatConvoGetMessages(ctx context.Context, convoId string, limit int) ([]*chat.ConvoDefs_MessageView, error) {

	var messages []*chat.ConvoDefs_MessageView
	cursor, lastId := "", ""

	// iterate until we got all records
	for {
		// query repo for collection with updated cursor
		output, err := chat.ConvoGetMessages(ctx, c.chatClient, convoId, cursor, 100)
		if err != nil {
			return nil, fmt.Errorf("ChatGetConvoMessages error: %v", err)
		}

		// stop if no records returned
		// or we get one we've already seen (maybe a repo doesn't support cursor?)
		if len(output.Messages) == 0 || lastId == output.Messages[len(output.Messages)-1].ConvoDefs_MessageView.Id {
			break
		}
		lastId = output.Messages[len(output.Messages)-1].ConvoDefs_MessageView.Id
		// store all record uris
		for _, msg := range output.Messages {
			messages = append(messages, msg.ConvoDefs_MessageView)
		}

		// if we have more records than the requested limit, stop
		// limit -1 indicates no upper limit, i.e. get all record
		if limit != -1 && len(messages) >= limit {
			break
		}
		// update cursor
		cursor = *output.Cursor
	}

	// don't return more than the requested limit
	var end int
	if limit == -1 {
		end = len(messages)
	} else {
		end = min(len(messages), limit)
	}
	return messages[:end], nil
}

// Send a message to the given account. Uses the existing chat with that account if it exists, or creates a new one if it doesn't.
func (c *Client) ChatSendMessage(ctx context.Context, handleOrDid string, message string) (string, string, error) {

	convo, err := c.ChatGetConvoForMembers(ctx, []string{handleOrDid})
	if err != nil {
		return "", "", err
	}

	return c.ChatConvoSendMessage(ctx, convo.Id, message)
}

// Send a group message to the given list of accounts. Uses the existing group chat with these accounts if it exists, or creates a new one if it doesn't.
func (c *Client) ChatSendGroupMessage(ctx context.Context, handlesOrDids []string, message string) (string, string, error) {
	convo, err := c.ChatGetConvoForMembers(ctx, handlesOrDids)
	if err != nil {
		return "", "", err
	}

	return c.ChatConvoSendMessage(ctx, convo.Id, message)
}

// Get all chat logs since the last cursor update (maintained internally).
func (c *Client) ChatGetRecentLogs(ctx context.Context) ([]*chat.ConvoGetLog_Output_Logs_Elem, error) {
	logOutput, err := chat.ConvoGetLog(ctx, c.chatClient, c.chatCursor)
	if err != nil {
		return nil, err
	}
	c.chatCursor = *logOutput.Cursor
	return logOutput.Logs, nil
}
