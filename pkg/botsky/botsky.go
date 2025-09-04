package botsky

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

const ApiEntryway = "https://bsky.social"
const ApiPublic = "https://public.api.bsky.app"
const ApiChat = "https://api.bsky.chat"

// TODO: need to wrap requests for rate limiting?

// API Client
//
// Wraps an XRPC client for API calls and a second one for handling chat/DMs
type Client struct {
	xrpcClient         *xrpc.Client
	Handle             string
	Did                string
	appkey             string
	refreshProcessLock sync.Mutex   // make sure only one auth refresher runs at a time
	chatClient         *xrpc.Client // client for accessing chat api
	chatCursor         string
}

// Sets up a new client (not yet authenticated)
func NewClient(ctx context.Context, handle string, appkey string) (*Client, error) {
	return NewClientWithPds(ctx, handle, appkey, ApiEntryway)
}
func NewClientWithPds(ctx context.Context, handle string, appkey string, server string) (*Client, error) {
	if server == "" {
		server = ApiEntryway
	}
	client := &Client{
		xrpcClient: &xrpc.Client{
			Client: new(http.Client),
			Host:   server,
		},
		Handle: handle,
		appkey: appkey,
		chatClient: &xrpc.Client{
			// TODO: reuse the http client?
			Client: new(http.Client),
			Host:   string(ApiChat),
		},
		chatCursor: "",
	}
	// resolve own handle to get did. don't need to be authenticated to do that
	clientDid, err := client.ResolveHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	client.Did = clientDid
	return client, nil
}

// Resolve the given handle to a DID
//
// If called on a DID, simply returns it
func (c *Client) ResolveHandle(ctx context.Context, handle string) (string, error) {
	if strings.HasPrefix(handle, "did:") {
		return handle, nil
	}
	if strings.HasPrefix(handle, "@") {
		handle = handle[1:]
	}
	output, err := atproto.IdentityResolveHandle(ctx, c.xrpcClient, handle)
	if err != nil {
		return "", fmt.Errorf("ResolveHandle error: %v", err)
	}
	return output.Did, nil
}

// Update the users profile description with the given string. All other profile components (avatar, banner, etc.) stay the same.
func (c *Client) UpdateProfileDescription(ctx context.Context, description string) error {
	profileRecord, err := atproto.RepoGetRecord(ctx, c.xrpcClient, "", "app.bsky.actor.profile", c.Handle, "self")
	if err != nil {
		return fmt.Errorf("UpdateProfileDescription error (RepoGetRecord): %v", err)
	}

	var actorProfile bsky.ActorProfile
	if err := decodeRecordAsLexicon(profileRecord.Value, &actorProfile); err != nil {
		return fmt.Errorf("UpdateProfileDescription error (DecodeRecordAsLexicon): %v", err)
	}

	newProfile := bsky.ActorProfile{
		LexiconTypeID:        "app.bsky.actor.profile",
		Avatar:               actorProfile.Avatar,
		Banner:               actorProfile.Banner,
		CreatedAt:            actorProfile.CreatedAt,
		Description:          &description,
		DisplayName:          actorProfile.DisplayName,
		JoinedViaStarterPack: actorProfile.JoinedViaStarterPack,
		Labels:               actorProfile.Labels,
		PinnedPost:           actorProfile.PinnedPost,
	}

	input := atproto.RepoPutRecord_Input{
		Collection: "app.bsky.actor.profile",
		Record: &lexutil.LexiconTypeDecoder{
			Val: &newProfile,
		},
		Repo:       c.Handle,
		Rkey:       "self",
		SwapRecord: profileRecord.Cid,
	}

	output, err := atproto.RepoPutRecord(ctx, c.xrpcClient, &input)
	if err != nil {
		return fmt.Errorf("UpdateProfileDescription error (RepoPutRecord): %v", err)
	}
	logger.Println("Profile updated:", output.Cid, output.Uri)
	return nil
}

// get posts from bsky API/AppView ***********************************************************

// TODO: method to get post directly from repo?
// Note: this fully relies on bsky api to be built

// Enriched post struct, including both the repo's FeedPost as well as bluesky's PostView
type RichPost struct {
	bsky.FeedPost

	AuthorDid   string // from *bsky.ActorDefs_ProfileViewBasic
	Cid         string
	Uri         string
	IndexedAt   string
	LikeCount   int64
	QuoteCount  int64
	ReplyCount  int64
	RepostCount int64

	Images []*bsky.EmbedImages_ViewImage
}

// Load Bluesky AppView postViews for the given repo/user.
//
// Set limit = -1 in order to get all postViews.
func (c *Client) GetPostViews(ctx context.Context, handleOrDid string, limit int) ([]*bsky.FeedDefs_PostView, error) {
	// get all post uris
	postUris, err := c.RepoGetRecordUris(ctx, handleOrDid, "app.bsky.feed.post", limit)
	if err != nil {
		return nil, fmt.Errorf("GetPostViews error (RepoGetRecordUris): %v", err)
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
			return nil, fmt.Errorf("GetPostViews error (FeedGetPosts): %v", err)
		}
		postViews = append(postViews, results.Posts...)
	}
	return postViews, nil
}

// Load enriched posts for repo/user.
//
// Set limit = -1 in order to get all posts.
func (c *Client) GetPosts(ctx context.Context, handleOrDid string, limit int) ([]*RichPost, error) {
	postViews, err := c.GetPostViews(ctx, handleOrDid, limit)
	if err != nil {
		return nil, fmt.Errorf("GetPosts error (GetPostViews): %v", err)
	}

	posts := make([]*RichPost, 0, len(postViews))
	for _, postView := range postViews {
		var feedPost bsky.FeedPost
		if err := decodeRecordAsLexicon(postView.Record, &feedPost); err != nil {
			return nil, fmt.Errorf("GetPosts error (DecodeRecordAsLexicon): %v", err)
		}
		posts = append(posts, &RichPost{
			FeedPost:    feedPost,
			AuthorDid:   postView.Author.Did,
			Cid:         postView.Cid,
			Uri:         postView.Uri,
			IndexedAt:   postView.IndexedAt,
			LikeCount:   *postView.LikeCount,
			QuoteCount:  *postView.QuoteCount,
			ReplyCount:  *postView.ReplyCount,
			RepostCount: *postView.RepostCount,
		})

	}
	return posts, nil
}

// Get a single post by uri.
func (c *Client) GetPost(ctx context.Context, postUri string) (RichPost, error) {
	results, err := bsky.FeedGetPosts(ctx, c.xrpcClient, []string{postUri})
	if err != nil {
		return RichPost{}, fmt.Errorf("GetPost error (FeedGetPosts): %v", err)
	}
	if len(results.Posts) == 0 {
		return RichPost{}, fmt.Errorf("GetPost error: No post with the given uri found")
	}
	postView := results.Posts[0]

	var feedPost bsky.FeedPost
	err = decodeRecordAsLexicon(postView.Record, &feedPost)
	if err != nil {
		return RichPost{}, fmt.Errorf("GetPost error (DecodeRecordAsLexicon): %v", err)
	}

	var images []*bsky.EmbedImages_ViewImage
	if postView.Embed != nil && postView.Embed.EmbedImages_View != nil {
		images = postView.Embed.EmbedImages_View.Images
	}

	post := RichPost{
		FeedPost:    feedPost,
		Images:      images,
		AuthorDid:   postView.Author.Did,
		Cid:         postView.Cid,
		Uri:         postView.Uri,
		IndexedAt:   postView.IndexedAt,
		LikeCount:   *postView.LikeCount,
		QuoteCount:  *postView.QuoteCount,
		ReplyCount:  *postView.ReplyCount,
		RepostCount: *postView.RepostCount,
	}

	return post, nil
}
