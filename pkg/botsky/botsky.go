package botsky

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

const DefaultServer = "https://bsky.social"

type Client struct {
	xrpcClient *xrpc.Client
	handle     string
	appkey     string
	// make sure only one auth refresher runs at a time
	refreshMutex sync.Mutex
}

// Sets up a new client connecting to the given server
func NewClient(ctx context.Context, server string, handle string, appkey string) (*Client, error) {
	client := &Client{
		xrpcClient: &xrpc.Client{
			Client: new(http.Client),
			Host:   server,
		},
		handle: handle,
		appkey: appkey,
	}
	return client, nil
}


func (c *Client) ResolveHandle(ctx context.Context, handle string) (string, error) {
	if strings.HasPrefix(handle, "@") {
		handle = handle[1:]
	}
	output, err := atproto.IdentityResolveHandle(ctx, c.xrpcClient, handle)
	if err != nil {
		return "", err
	}
	return output.Did, nil
}

// get posts from bsky API/AppView ***********************************************************

// Enriched post struct, including both the repo's FeedPost as well as bluesky's PostView
// Note: this fully relies on bsky api to be built
type RichPost struct {
	bsky.FeedPost

	AuthorDid   string // from *bsky.ActorDefs_ProfileViewBasic
	Cid         string
	Uri         string
	IndexedAt   string
	Labels      []*atproto.LabelDefs_Label
	LikeCount   int64
	QuoteCount  int64
	ReplyCount  int64
	RepostCount int64
}

// Load bsky postViews from repo/user
func (c *Client) GetPostViews(ctx context.Context, handleOrDid string, limit int) ([]*bsky.FeedDefs_PostView, error) {
	// get all post uris
	postUris, err := c.RepoGetRecordUris(ctx, handleOrDid, "app.bsky.feed.post", limit)
	if err != nil {
		return nil, err
	}

	// hydrate'em
	postViews := make([]*bsky.FeedDefs_PostView, 0, len(postUris))
	for i := 0; i < len(postUris); i += 25 {
		j := i + 25
		if j > len(postUris) {
			j = len(postUris)
		}
		results, err := bsky.FeedGetPosts(ctx, c.xrpcClient, postUris[i:j])
		if err != nil {
			return nil, err
		}
		postViews = append(postViews, results.Posts...)
	}
	return postViews, nil
}

// Load enriched posts from repo/user
func (c *Client) GetPosts(ctx context.Context, handleOrDid string, limit int) ([]*RichPost, error) {
	postViews, err := c.GetPostViews(ctx, handleOrDid, limit)
	if err != nil {
		return nil, err
	}

	posts := make([]*RichPost, 0, len(postViews))
	for _, postView := range postViews {
		var feedPost bsky.FeedPost
		if err := DecodeRecordAsLexicon(postView.Record, &feedPost); err != nil {
			fmt.Println("failed to decode postView.Record as FeedPost:", err)
		} else {
			posts = append(posts, &RichPost{
				FeedPost:    feedPost,
				AuthorDid:   postView.Author.Did,
				Cid:         postView.Cid,
				Uri:         postView.Uri,
				IndexedAt:   postView.IndexedAt,
				Labels:      postView.Labels,
				LikeCount:   *postView.LikeCount,
				QuoteCount:  *postView.QuoteCount,
				ReplyCount:  *postView.ReplyCount,
				RepostCount: *postView.RepostCount,
			})
		}
	}
	return posts, nil
}

func (c *Client) GetPost(ctx context.Context, postUri string) (RichPost, error) {
	results, err := bsky.FeedGetPosts(ctx, c.xrpcClient, []string{postUri})
    if err != nil {
        return RichPost{}, fmt.Errorf("Unable to get feedpost for given postUri: %v", err)
    }
    if len(results.Posts) == 0 {
        return RichPost{}, fmt.Errorf("No post with the given uri found")
    }
    postView := results.Posts[0]

    var feedPost bsky.FeedPost
    err = DecodeRecordAsLexicon(postView.Record, &feedPost)
    if err != nil {
        return RichPost{}, fmt.Errorf("Unable to decode FeedPost from PostView.")
    }

    post := RichPost{
        FeedPost: feedPost,        
        AuthorDid:   postView.Author.Did,
        Cid:         postView.Cid,
        Uri:         postView.Uri,
        IndexedAt:   postView.IndexedAt,
        Labels:      postView.Labels,
        LikeCount:   *postView.LikeCount,
        QuoteCount:  *postView.QuoteCount,
        ReplyCount:  *postView.ReplyCount,
        RepostCount: *postView.RepostCount,
    }

    return post, nil
}
