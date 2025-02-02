// This file has been heavily modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/post.go
// License: Apache 2.0

package botsky

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/davhofer/indigo/api/atproto"
	"github.com/davhofer/indigo/api/bsky"
	lexutil "github.com/davhofer/indigo/lex/util"
)

// TODO: embed videos

type Facet_Type int

const (
	Facet_Link Facet_Type = iota + 1
	Facet_Mention
	Facet_Tag
)


type InlineLink struct {
	Text string
	Url  string
}

type Facet struct {
	Ftype   Facet_Type
	Value   string
	T_facet string
}

type RecordRef struct {
    Cid string 
    Uri string
}


type Embed struct {
	Link           Link
	Images         []Image
	UploadedImages []lexutil.LexBlob
    Record         RecordRef
}

type ReplyReference struct {
    Uri string 
    Cid string 
    RootUri string 
    RootCid string
}


type Link struct {
	Title       string
	Uri         url.URL
	Description string
	Thumb       lexutil.LexBlob
}

type ImageSource struct {
	Alt string
	Uri string
}
type Image struct {
	Alt string
	Uri url.URL
}


func (c *Client) Repost(ctx context.Context, postUri string) (string, string, error) {

    cid, _, err := c.RepoGetPost(ctx, postUri)
    if err != nil {
        return "", "", fmt.Errorf("Error getting post to repost: %v", err)
    }
    ref := atproto.RepoStrongRef{
        Uri: postUri,
        Cid: cid,
    }

    post := bsky.FeedRepost{
        LexiconTypeID: "app.bsky.feed.repost", 
        CreatedAt: time.Now().Format(time.RFC3339),
        Subject: &ref,
    }
	
	post_input := &atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.repost",
		Repo: c.XrpcClient.Auth.Did,
		Record: &lexutil.LexiconTypeDecoder{Val: &post},
	}
	response, err := atproto.RepoCreateRecord(ctx, c.XrpcClient, post_input)
	if err != nil {
        return "", "", fmt.Errorf("unable to repost: %v", err)
	}

    return response.Cid, response.Uri, nil
}

type PostBuilder struct {
	Text  string
	Facet []Facet
    ReplyUri string
    ReplyReference ReplyReference
	Embed Embed
    EmbedLink string 
    EmbedImages []ImageSource
    EmbedPostQuote string
    AdditionalTags []string
    HasEmbed bool
    Languages []string
    RenderHashtags bool
    Mentions []string
}

func NewPostBuilder(text string) *PostBuilder {
    pb := &PostBuilder{
		Text:  text,
        RenderHashtags: true,
	}

    return pb
}

func (pb *PostBuilder) AddTags(tags []string) *PostBuilder {
    pb.AdditionalTags = append(pb.AdditionalTags, tags...)
    return pb
}

func (pb *PostBuilder) AddFacet(ftype Facet_Type, value string, text string) *PostBuilder {
	pb.Facet = append(pb.Facet, Facet{
		Ftype:   ftype,
		Value:   value,
		T_facet: text,
	})
    return pb
}

func (pb *PostBuilder) AddInlineLinks(links []InlineLink) *PostBuilder {
    for _, link := range links {
        pb.AddFacet(Facet_Link, link.Url, link.Text)
    }
    return pb
}


func (pb *PostBuilder) AddLanguage(language string) *PostBuilder {
    pb.Languages = append(pb.Languages, language)
    return pb
}

func (pb *PostBuilder) AddMentions(mentions []string) *PostBuilder {
    pb.Mentions = append(pb.Mentions, mentions...)
    return pb
}

func (pb *PostBuilder) ReplyTo(postUri string) *PostBuilder {
    pb.ReplyUri = postUri
    return pb
}

func (pb *PostBuilder) AddEmbedLink(link string) *PostBuilder {
    pb.EmbedLink = link
    return pb
}

func (pb *PostBuilder) AddImages(images []ImageSource) *PostBuilder {
    pb.EmbedImages = append(pb.EmbedImages, images...)
    return pb
}

func (pb *PostBuilder) AddQuotedPost(postUri string) *PostBuilder {
    pb.EmbedPostQuote = postUri
    return pb
}


func (c *Client) Post(ctx context.Context, pb *PostBuilder) (string, string, error) {
    nEmbeds := 0 
	if pb.EmbedImages != nil {
        nEmbeds++
    }
    if pb.EmbedLink != "" {
        nEmbeds++
	}
    if pb.EmbedPostQuote != "" {
        nEmbeds++ 
    }
	
    if nEmbeds > 1 {
        return "", "", fmt.Errorf("Can only include one type of Embed (images, embedded link, quoted post) in posts.")
    }

	if len(pb.Languages) == 0 {
		pb.Languages = []string{"en"}
	}

	if pb.Mentions != nil {
		for _, handle := range pb.Mentions {
			var resolveHandle string
			if strings.HasPrefix(handle, "@") {
				resolveHandle = handle[1:]
			} else {
				resolveHandle = handle
			}
			resolveOutput, err := atproto.IdentityResolveHandle(ctx, c.XrpcClient, resolveHandle)
			if err != nil {
                return "", "", fmt.Errorf("Unable to resolve handle: %v", err)
			}
			pb.AddFacet(Facet_Mention, resolveOutput.Did, handle)
		}

	}

	if pb.EmbedImages != nil {
		var parsedImages []Image
		for _, img := range pb.EmbedImages {
			parsedUrl, err := url.Parse(img.Uri)
			if err != nil {
				log.Println("Unable to parse image source uri", img.Uri)
			} else {
				parsedImages = append(parsedImages, Image{Alt: img.Alt, Uri: *parsedUrl})
			}
		}
		if len(parsedImages) > 0 {
			blobs, err := c.RepoUploadImages(ctx, parsedImages)
			if err != nil {
                return "", "", fmt.Errorf("Error when uploading images: %v", err)
			}
            pb.Embed.Images = parsedImages
            pb.Embed.UploadedImages = blobs
		}
	}

	if pb.EmbedLink != "" {
		parsedLink, err := url.Parse(pb.EmbedLink)
		if err != nil {
            return "", "", fmt.Errorf("Error when parsing link: %v", err)
		}

		siteTags, err := fetchOpenGraphTwitterTags(pb.EmbedLink)
		if err != nil {
            return "", "", fmt.Errorf("Error when fetching og/twitter tags from link: %v", err)
		}

		title := siteTags["title"]
		description := siteTags["description"]
		imageUrl, hasImage := siteTags["image"]
		alt := siteTags["image:alt"]

		var blob lexutil.LexBlob
		if hasImage {
			parsedImageUrl, err := url.Parse(imageUrl)
			if err != nil {
                return "", "", fmt.Errorf("Error when parsing image url: %v", err)
			}
			previewImg := Image{
				Uri: *parsedImageUrl,
				Alt: alt,
			}
			b, err := c.RepoUploadImage(ctx, previewImg)
			if err != nil {
                return "", "", fmt.Errorf("Error when trying to upload image: %v", err)
			}
			if b != nil {
				blob = *b
			}
		}

        pb.Embed.Link.Title = title
        pb.Embed.Link.Uri = *parsedLink
        pb.Embed.Link.Description = description
        pb.Embed.Link.Thumb = blob
	}

    if pb.EmbedPostQuote != "" {
        cid, _, err := c.RepoGetPost(ctx, pb.EmbedPostQuote) 
        if err != nil {
            return "", "", fmt.Errorf("Error when getting quoted post: %v", err)
        }
        pb.Embed.Record.Cid = cid 
        pb.Embed.Record.Uri = pb.EmbedPostQuote
    }

    if pb.ReplyUri != "" {
        cid, replyPost, err := c.RepoGetPost(ctx, pb.ReplyUri) 
        if err != nil {
            return "", "", fmt.Errorf("Error when getting reply post: %v", err)
        }
        
        var rootCid, rootUri string
        if replyPost.Reply != nil && *replyPost.Reply != (bsky.FeedPost_ReplyRef{}) {
            rootCid = replyPost.Reply.Root.Cid
            rootUri = replyPost.Reply.Root.Uri
        } else {
            rootCid = cid
            rootUri = pb.ReplyUri
        }

        pb.ReplyReference = ReplyReference{
            Uri: pb.ReplyUri,
            Cid: cid,
            RootUri: rootUri,
            RootCid: rootCid,
        }
    }

	// Build post
	post, err := pb.Build()
	if err != nil {
        return "", "", fmt.Errorf("Error when building post: %v", err)
	}

	return c.RepoCreatePostRecord(ctx, post)

}

// Build the request
func (pb *PostBuilder) Build() (bsky.FeedPost, error) {

	post := bsky.FeedPost{Langs: pb.Languages}

	post.Text = pb.Text
	post.LexiconTypeID = "app.bsky.feed.post"
	post.CreatedAt = time.Now().Format(time.RFC3339)
    post.Tags = pb.AdditionalTags

	// RichtextFacet Section
	// https://docs.bsky.app/docs/advanced-guides/post-richtext

	Facets := []*bsky.RichtextFacet{}

	for _, f := range pb.Facet {
		facet := &bsky.RichtextFacet{}
		features := []*bsky.RichtextFacet_Features_Elem{}
		feature := &bsky.RichtextFacet_Features_Elem{}

		switch f.Ftype {

		case Facet_Link:
			{
				feature = &bsky.RichtextFacet_Features_Elem{
					RichtextFacet_Link: &bsky.RichtextFacet_Link{
						LexiconTypeID: f.Ftype.String(),
						Uri:           f.Value,
					},
				}
			}

		case Facet_Mention:
			{
				feature = &bsky.RichtextFacet_Features_Elem{
					RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
						LexiconTypeID: f.Ftype.String(),
						Did:           f.Value,
					},
				}
			}

		case Facet_Tag:
			{
				feature = &bsky.RichtextFacet_Features_Elem{
					RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
						LexiconTypeID: f.Ftype.String(),
						Tag:           f.Value,
					},
				}
			}

		}

		features = append(features, feature)
		facet.Features = features

		ByteStart, ByteEnd, err := findSubstring(post.Text, f.T_facet)
		if err != nil {
			return post, fmt.Errorf("unable to find the substring: %v , %v", f.T_facet, err)
		}

		index := &bsky.RichtextFacet_ByteSlice{
			ByteStart: int64(ByteStart),
			ByteEnd:   int64(ByteEnd),
		}
		facet.Index = index

		Facets = append(Facets, facet)
	}

	// We parse hashtags with regex instead of relying on substring matching
	// The reason is that it is relatively common to have similar/overalpping hashtags, like
	// #atproto and #atprotodev, which could lead to mistakes
	if pb.RenderHashtags {
		hashtagRegex := `(?:^|\s)(#[^\d\s]\S*)`
		matches := findRegexMatches(post.Text, hashtagRegex)
		for _, m := range matches {
			facet := &bsky.RichtextFacet{}
			features := []*bsky.RichtextFacet_Features_Elem{}
			feature := &bsky.RichtextFacet_Features_Elem{}

			feature = &bsky.RichtextFacet_Features_Elem{
				RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
					LexiconTypeID: Facet_Tag.String(),
					Tag:           stripHashtag(m.Value),
				},
			}

			features = append(features, feature)
			facet.Features = features

			index := &bsky.RichtextFacet_ByteSlice{
				ByteStart: int64(m.Start),
				ByteEnd:   int64(m.End),
			}
			facet.Index = index

			Facets = append(Facets, facet)
		}
	}

	post.Facets = Facets

	var FeedPost_Embed bsky.FeedPost_Embed
    embedFlag := true

	// Embed Section (either external links or images)
	// As of now it allows only one Embed type per post:
	// https://github.com/bluesky-social/indigo/blob/main/api/bsky/feedpost.go
	if pb.Embed.Link != (Link{}) {

		FeedPost_Embed.EmbedExternal = &bsky.EmbedExternal{
			LexiconTypeID: "app.bsky.embed.external",
			External: &bsky.EmbedExternal_External{
				Title:       pb.Embed.Link.Title,
				Uri:         pb.Embed.Link.Uri.String(),
				Description: pb.Embed.Link.Description,
				Thumb:       &pb.Embed.Link.Thumb,
			},
		}

	} else if len(pb.Embed.Images) != 0 && len(pb.Embed.Images) == len(pb.Embed.UploadedImages) {

		EmbedImages := bsky.EmbedImages{
			LexiconTypeID: "app.bsky.embed.images",
			Images:        make([]*bsky.EmbedImages_Image, len(pb.Embed.Images)),
		}

		for i, img := range pb.Embed.Images {
			EmbedImages.Images[i] = &bsky.EmbedImages_Image{
				Alt:   img.Alt,
				Image: &pb.Embed.UploadedImages[i],
			}
		}

		FeedPost_Embed.EmbedImages = &EmbedImages

	} else if pb.Embed.Record != (RecordRef{}) {
        EmbedRecord := bsky.EmbedRecord{
            LexiconTypeID: "app.bsky.embed.record",
            Record: &atproto.RepoStrongRef{
                LexiconTypeID: "com.atproto.repo.strongRef", 
                Cid: pb.Embed.Record.Cid,
                Uri: pb.Embed.Record.Uri,
            },
        }

        FeedPost_Embed.EmbedRecord = &EmbedRecord
    } else {
        embedFlag = false 
    }

	// avoid error when trying to marshal empty field (*bsky.FeedPost_Embed)
	if embedFlag {
		post.Embed = &FeedPost_Embed
	}


    // set reply 
    if pb.ReplyReference != (ReplyReference{}) {
        post.Reply = &bsky.FeedPost_ReplyRef{
            Parent: &atproto.RepoStrongRef{
                Uri: pb.ReplyReference.Uri,
                Cid: pb.ReplyReference.Cid,
            },
            Root: &atproto.RepoStrongRef{
                Uri: pb.ReplyReference.RootUri,
                Cid: pb.ReplyReference.RootCid,
            },
        }
    }

	return post, nil
}

func (f Facet_Type) String() string {
	switch f {
	case Facet_Link:
		return "app.bsky.richtext.facet#link"
	case Facet_Mention:
		return "app.bsky.richtext.facet#mention"
	case Facet_Tag:
		return "app.bsky.richtext.facet#tag"
	default:
		return "Unknown"
	}
}
