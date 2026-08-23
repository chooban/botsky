package botsky

import (
	"net/url"
	"testing"
	"time"

	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseURL(t *testing.T, raw string) url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return *u
}

func lexBlobForTest(t *testing.T) lexutil.LexBlob {
	t.Helper()
	return lexutil.LexBlob{
		MimeType: "image/jpeg",
		Size:     1234,
	}
}

func TestFacetTypeString(t *testing.T) {
	assert.Equal(t, "app.bsky.richtext.facet#link", facetTypeLink.String())
	assert.Equal(t, "app.bsky.richtext.facet#mention", facetTypeMention.String())
	assert.Equal(t, "app.bsky.richtext.facet#tag", facetTypeTag.String())
	assert.Equal(t, "Unknown", facetType(0).String())
}

func TestBuildPostPlainText(t *testing.T) {
	pb := NewPostBuilder("Hello world").AddTags([]string{"indieweb", "blogging"})
	pb.Languages = []string{"en"}

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	assert.Equal(t, "app.bsky.feed.post", post.LexiconTypeID)
	assert.Equal(t, "Hello world", post.Text)
	assert.Equal(t, []string{"indieweb", "blogging"}, post.Tags)
	assert.Equal(t, []string{"en"}, post.Langs)
	assert.Empty(t, post.Facets)
	assert.Nil(t, post.Embed)
	assert.Nil(t, post.Reply)

	_, err = time.Parse(time.RFC3339, post.CreatedAt)
	require.NoError(t, err)
}

func TestBuildPostMentionFacet(t *testing.T) {
	pb := NewPostBuilder("hello @alice.example.com")
	matches := []struct {
		Value string
		Start int
		End   int
		Did   string
	}{
		{Value: "alice.example.com", Start: 6, End: 24, Did: "did:plc:alice"},
	}

	post, err := buildPost(pb, embed{}, replyReference{}, matches)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Equal(t, int64(6), facet.Index.ByteStart)
	require.Equal(t, int64(24), facet.Index.ByteEnd)
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Mention)
	assert.Equal(t, "app.bsky.richtext.facet#mention", facet.Features[0].RichtextFacet_Mention.LexiconTypeID)
	assert.Equal(t, "did:plc:alice", facet.Features[0].RichtextFacet_Mention.Did)
}

func TestBuildPostInlineLinkFacet(t *testing.T) {
	pb := NewPostBuilder("Check out my blog example.com for more")
	pb.AddInlineLinks([]InlineLink{{Text: "blog", Url: "https://example.com/"}})

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Equal(t, int64(13), facet.Index.ByteStart)
	require.Equal(t, int64(17), facet.Index.ByteEnd)
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Link)
	assert.Equal(t, "app.bsky.richtext.facet#link", facet.Features[0].RichtextFacet_Link.LexiconTypeID)
	assert.Equal(t, "https://example.com/", facet.Features[0].RichtextFacet_Link.Uri)
}

func TestBuildPostInlineLinkNotFound(t *testing.T) {
	pb := NewPostBuilder("no link text here")
	pb.AddInlineLinks([]InlineLink{{Text: "missing", Url: "https://example.com/"}})

	_, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.Error(t, err)
}

func TestBuildPostAutoDetectedLinkFacet(t *testing.T) {
	pb := NewPostBuilder("See https://example.com for details")

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Equal(t, int64(4), facet.Index.ByteStart)
	require.Equal(t, int64(23), facet.Index.ByteEnd)
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Link)
	assert.Equal(t, "https://example.com", facet.Features[0].RichtextFacet_Link.Uri)
}

func TestBuildPostAutoDetectedLinkFacetWithPath(t *testing.T) {
	pb := NewPostBuilder("See https://example.com/foo for details")

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Link)
	assert.Equal(t, "https://example.com/foo", facet.Features[0].RichtextFacet_Link.Uri)
	// the facet covers the whole URL
	assert.Equal(t, int64(4), facet.Index.ByteStart)
	assert.Equal(t, int64(4+len("https://example.com/foo")), facet.Index.ByteEnd)
}

func TestBuildPostAutoDetectedLinkFacetWithQuery(t *testing.T) {
	pb := NewPostBuilder("See https://example.com/search?q=go#top for details")

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Link)
	assert.Equal(t, "https://example.com/search?q=go#top", facet.Features[0].RichtextFacet_Link.Uri)
	assert.Equal(t, int64(4), facet.Index.ByteStart)
	assert.Equal(t, int64(4+len("https://example.com/search?q=go#top")), facet.Index.ByteEnd)
}

func TestBuildPostAutoDetectedLinkFacetTrimsPunctuation(t *testing.T) {
	pb := NewPostBuilder("See https://example.com/foo, and https://example.com/bar.")

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 2)

	first := post.Facets[0].Features[0].RichtextFacet_Link
	require.NotNil(t, first)
	assert.Equal(t, "https://example.com/foo", first.Uri)
	assert.Equal(t, int64(4), post.Facets[0].Index.ByteStart)
	assert.Equal(t, int64(4+len("https://example.com/foo")), post.Facets[0].Index.ByteEnd)

	second := post.Facets[1].Features[0].RichtextFacet_Link
	require.NotNil(t, second)
	assert.Equal(t, "https://example.com/bar", second.Uri)
	assert.Equal(t, int64(33), post.Facets[1].Index.ByteStart)
	assert.Equal(t, int64(33+len("https://example.com/bar")), post.Facets[1].Index.ByteEnd)
}

func TestBuildPostAutoDetectedLinkFacetKeepsBalancedBrackets(t *testing.T) {
	pb := NewPostBuilder("See https://en.wikipedia.org/wiki/Example_(disambiguation) for details")

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Link)
	assert.Equal(t, "https://en.wikipedia.org/wiki/Example_(disambiguation)", facet.Features[0].RichtextFacet_Link.Uri)
}

func TestBuildPostHashtagFacet(t *testing.T) {
	pb := NewPostBuilder("Hello #indieweb world")

	post, err := buildPost(pb, embed{}, replyReference{}, nil)
	require.NoError(t, err)

	require.Len(t, post.Facets, 1)
	facet := post.Facets[0]
	require.Equal(t, int64(5), facet.Index.ByteStart)
	require.Equal(t, int64(15), facet.Index.ByteEnd)
	require.Len(t, facet.Features, 1)
	require.NotNil(t, facet.Features[0].RichtextFacet_Tag)
	assert.Equal(t, "app.bsky.richtext.facet#tag", facet.Features[0].RichtextFacet_Tag.LexiconTypeID)
	assert.Equal(t, "indieweb", facet.Features[0].RichtextFacet_Tag.Tag)
}

func TestBuildPostExternalEmbed(t *testing.T) {
	pb := NewPostBuilder("Check this out")

	post, err := buildPost(pb, embed{
		Link: embedLink{
			Title:       "An article",
			Uri:         mustParseURL(t, "https://example.com/article"),
			Description: "A description of the article",
		},
	}, replyReference{}, nil)
	require.NoError(t, err)

	require.NotNil(t, post.Embed)
	require.NotNil(t, post.Embed.EmbedExternal)
	assert.Equal(t, "app.bsky.embed.external", post.Embed.EmbedExternal.LexiconTypeID)
	require.NotNil(t, post.Embed.EmbedExternal.External)
	assert.Equal(t, "An article", post.Embed.EmbedExternal.External.Title)
	assert.Equal(t, "https://example.com/article", post.Embed.EmbedExternal.External.Uri)
	assert.Equal(t, "A description of the article", post.Embed.EmbedExternal.External.Description)
}

func TestBuildPostImagesEmbed(t *testing.T) {
	pb := NewPostBuilder("Some photos")
	blob := lexBlobForTest(t)

	post, err := buildPost(pb, embed{
		Images: []imageSourceParsed{
			{Alt: "first", Uri: mustParseURL(t, "https://example.com/one.jpg")},
			{Alt: "second", Uri: mustParseURL(t, "https://example.com/two.jpg")},
		},
		AspectRatios:   []*ImageDimensions{{Width: 100, Height: 50}, nil},
		UploadedImages: []lexutil.LexBlob{blob, blob},
	}, replyReference{}, nil)
	require.NoError(t, err)

	require.NotNil(t, post.Embed)
	require.NotNil(t, post.Embed.EmbedImages)
	assert.Equal(t, "app.bsky.embed.images", post.Embed.EmbedImages.LexiconTypeID)
	require.Len(t, post.Embed.EmbedImages.Images, 2)

	first := post.Embed.EmbedImages.Images[0]
	assert.Equal(t, "first", first.Alt)
	assert.Equal(t, &blob, first.Image)
	require.NotNil(t, first.AspectRatio)
	assert.Equal(t, int64(100), first.AspectRatio.Width)
	assert.Equal(t, int64(50), first.AspectRatio.Height)

	second := post.Embed.EmbedImages.Images[1]
	assert.Equal(t, "second", second.Alt)
	assert.Nil(t, second.AspectRatio)
}

func TestBuildPostQuotedRecordEmbed(t *testing.T) {
	pb := NewPostBuilder("Great point")

	post, err := buildPost(pb, embed{
		Record: recordRef{
			Cid: "bafyreibvjvcv45gazq4xqv6l7mhx5u4rh2s2z4d4d4",
			Uri: "at://did:plc:alice/app.bsky.feed.post/abc123",
		},
	}, replyReference{}, nil)
	require.NoError(t, err)

	require.NotNil(t, post.Embed)
	require.NotNil(t, post.Embed.EmbedRecord)
	assert.Equal(t, "app.bsky.embed.record", post.Embed.EmbedRecord.LexiconTypeID)
	require.NotNil(t, post.Embed.EmbedRecord.Record)
	assert.Equal(t, "com.atproto.repo.strongRef", post.Embed.EmbedRecord.Record.LexiconTypeID)
	assert.Equal(t, "at://did:plc:alice/app.bsky.feed.post/abc123", post.Embed.EmbedRecord.Record.Uri)
	assert.Equal(t, "bafyreibvjvcv45gazq4xqv6l7mhx5u4rh2s2z4d4d4", post.Embed.EmbedRecord.Record.Cid)
}

func TestBuildPostReplyRef(t *testing.T) {
	pb := NewPostBuilder("A reply")
	pb.ReplyTo("at://did:plc:alice/app.bsky.feed.post/parent")

	post, err := buildPost(pb, embed{}, replyReference{
		Uri:     "at://did:plc:alice/app.bsky.feed.post/parent",
		Cid:     "parent-cid",
		RootUri: "at://did:plc:alice/app.bsky.feed.post/root",
		RootCid: "root-cid",
	}, nil)
	require.NoError(t, err)

	require.NotNil(t, post.Reply)
	require.NotNil(t, post.Reply.Parent)
	assert.Equal(t, "at://did:plc:alice/app.bsky.feed.post/parent", post.Reply.Parent.Uri)
	assert.Equal(t, "parent-cid", post.Reply.Parent.Cid)
	require.NotNil(t, post.Reply.Root)
	assert.Equal(t, "at://did:plc:alice/app.bsky.feed.post/root", post.Reply.Root.Uri)
	assert.Equal(t, "root-cid", post.Reply.Root.Cid)
}

func TestBuildPostFacetsCombine(t *testing.T) {
	pb := NewPostBuilder("hello @alice.example.com #indieweb https://example.com")
	matches := []struct {
		Value string
		Start int
		End   int
		Did   string
	}{
		{Value: "alice.example.com", Start: 6, End: 24, Did: "did:plc:alice"},
	}

	post, err := buildPost(pb, embed{}, replyReference{}, matches)
	require.NoError(t, err)

	// mention + hashtag + auto-detected link
	require.Len(t, post.Facets, 3)
}
