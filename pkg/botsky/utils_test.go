package botsky

import (
	"testing"

	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindSubstring(t *testing.T) {
	start, end, err := findSubstring("hello world", "world")
	require.NoError(t, err)
	assert.Equal(t, 6, start)
	assert.Equal(t, 11, end)

	// only the first occurrence is returned
	start, end, err = findSubstring("abc abc", "abc")
	require.NoError(t, err)
	assert.Equal(t, 0, start)
	assert.Equal(t, 3, end)

	_, _, err = findSubstring("hello world", "missing")
	require.Error(t, err)

	start, end, err = findSubstring("hello world", "")
	require.NoError(t, err)
	assert.Equal(t, 0, start)
	assert.Equal(t, 0, end)
}

func TestFindRegexMatches(t *testing.T) {
	matches := findRegexMatches("a1 b2 c3", `[a-z][0-9]`)
	require.Len(t, matches, 3)
	assert.Equal(t, "a1", matches[0].Value)
	assert.Equal(t, 0, matches[0].Start)
	assert.Equal(t, 2, matches[0].End)
	assert.Equal(t, "c3", matches[2].Value)
	assert.Equal(t, 6, matches[2].Start)
	assert.Equal(t, 8, matches[2].End)

	// anchored pattern only matches at the start
	matches = findRegexMatches("a1 b2", `^[a-z]`)
	require.Len(t, matches, 1)

	// no matches
	matches = findRegexMatches("123", `[a-z]`)
	assert.Empty(t, matches)
}

func TestStripHashtag(t *testing.T) {
	assert.Equal(t, "indieweb", stripHashtag("#indieweb"))
	assert.Equal(t, "foo", stripHashtag("#foo."))
	assert.Equal(t, "bar", stripHashtag(" #bar "))
	assert.Equal(t, "plain", stripHashtag("plain"))
	assert.Equal(t, "123abc", stripHashtag("#123abc"))
	// '+' is a math symbol, not punctuation, so it is preserved
	assert.Equal(t, "C++", stripHashtag("#C++"))
}

func TestTrimURLTrailing(t *testing.T) {
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo."))
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo,"))
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo;"))
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo?"))
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo"))
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo)"))
	assert.Equal(t, "https://example.com/foo/bar", trimURLTrailing("https://example.com/foo/bar."))
	// closing brackets are kept when they are balanced
	assert.Equal(t, "https://en.wikipedia.org/wiki/Example_(disambiguation)", trimURLTrailing("https://en.wikipedia.org/wiki/Example_(disambiguation)"))
	assert.Equal(t, "https://example.com/foo", trimURLTrailing("https://example.com/foo))"))
}

var png1x1 = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0x63, 0x64, 0x60, 0xF8, 0x5F,
	0x0F, 0x00, 0x02, 0x87, 0x01, 0x80, 0xEB, 0x47, 0xBA, 0x92, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
	0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestGetImageDimensions(t *testing.T) {
	dims := getImageDimensions(png1x1)
	require.NotNil(t, dims)
	assert.Equal(t, 1, dims.Width)
	assert.Equal(t, 1, dims.Height)

	assert.Nil(t, getImageDimensions([]byte("not an image")))
	assert.Nil(t, getImageDimensions(nil))
}

func TestDecodeRecordAsLexicon(t *testing.T) {
	input := &bsky.FeedPost{
		LexiconTypeID: "app.bsky.feed.post",
		Text:          "hello @alice.example.com",
		Langs:         []string{"en"},
		Tags:          []string{"indieweb"},
		CreatedAt:     "2024-01-01T00:00:00Z",
		Facets: []*bsky.RichtextFacet{
			{
				Index: &bsky.RichtextFacet_ByteSlice{ByteStart: 6, ByteEnd: 24},
				Features: []*bsky.RichtextFacet_Features_Elem{
					{
						RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
							LexiconTypeID: "app.bsky.richtext.facet#mention",
							Did:           "did:plc:alice",
						},
					},
				},
			},
		},
	}

	decoder := &lexutil.LexiconTypeDecoder{Val: input}

	var decoded bsky.FeedPost
	err := decodeRecordAsLexicon(decoder, &decoded)
	require.NoError(t, err)
	require.Equal(t, *input, decoded)
}
