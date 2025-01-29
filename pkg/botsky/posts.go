// This file has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/post.go
// License: Apache 2.0

package botsky

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
)

// TODO: reply posts, quote posts, website card embeds? => maybe directly use website screenshotter, or how other sites implement that feature?




type Facet_Type int

const (
	Facet_Link Facet_Type = iota + 1
	Facet_Mention
	Facet_Tag
)

type PostBuilder struct {
	Text  string
	Facet []Facet
	Embed Embed
}

type InlineLink struct {
    Text string 
    Url string
}

type Facet struct {
	Ftype   Facet_Type
	Value   string
	T_facet string
}

type Embed struct {
	Link           Link
	Images         []Image
	UploadedImages []lexutil.LexBlob
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
	Uri   url.URL
}

/*
(check: where can we have one, where can we have many?)
stuff we can have in our post:
- text
- facets:
    - link => points into text
    - mentions => points into text
    - tags 
- external link: embedded link with preview 
- images


for mentions: let user give the handles of mentioned people, automatically resolve using com.atproto.identity.resolveHandle
for links: let user provide substring and actual url?

TODO: mentions => manual (give list of people to mention, referring to text)
TODO: hashtags => detect automatically... if we just give list of tags, and one is a substring of the other, we can run into problems...
*/
func (c *Client) NewPost(ctx context.Context, text string, mentions []string, inlineLinks []InlineLink, images []ImageSource, embeddedLink string, renderHashtags bool, languages []string) (string, string, error) {
    if images != nil && embeddedLink != "" {
        return "", "", fmt.Errorf("Can only include one type of Embed (images or embedded link) in posts.")
    }

    if len(languages) == 0 {
        languages = []string{ "en" }
    }

    // Set up PostBuilder
    pb := NewPostBuilder(text)

    if mentions != nil {
        for _, handle := range mentions {
            var resolveHandle string
            if strings.HasPrefix(handle, "@") {
                resolveHandle = handle[1:]
            } else {
                resolveHandle = handle
            }
            resolveOutput, err := atproto.IdentityResolveHandle(ctx, c.xrpcClient, resolveHandle)
            if err != nil {
                return "", "", err 
            }
            pb = pb.WithFacet(Facet_Mention, resolveOutput.Did, handle)
        }

    }

    if inlineLinks != nil {
        for _, link := range inlineLinks {
            pb = pb.WithFacet(Facet_Link, link.Url, link.Text)
        }

    }

    if images != nil {
        var parsedImages []Image
        for _, img := range images {
            parsedUrl, err := url.Parse(img.Uri)
            if err != nil {
                log.Println("Unable to parse image source uri", img.Uri)
            } else {
                parsedImages = append(parsedImages, Image{Alt: img.Alt, Uri: *parsedUrl})
            }
        }
        if len(parsedImages) > 0 {
            blobs, err := c.UploadImages(ctx, parsedImages)
            if err != nil {
                return "", "", err
            }
            pb = pb.WithImages(blobs, parsedImages)
        }
    }

    if embeddedLink != "" {
        parsedLink, err := url.Parse(embeddedLink)
        if err != nil {
            return "", "", err
        }

        siteTags, err := fetchOpenGraphTwitterTags(embeddedLink)
        if err != nil {
            return "", "", err
        }

        title := siteTags["title"]
        description := siteTags["description"]
        imageUrl, hasImage := siteTags["image"]
        alt := siteTags["image:alt"]

        var blob lexutil.LexBlob
        if hasImage {
            parsedImageUrl, err := url.Parse(imageUrl)
            if err != nil {
                return "", "", err
            }
            previewImg := Image{
                Uri: *parsedImageUrl,
                Alt: alt,
            }
            b, err := c.UploadImage(ctx, previewImg)
            if err != nil {
                return "", "", err
            }
            if b != nil {
                blob = *b 
            }
        }

        pb = pb.WithExternalLink(title, *parsedLink, description, blob)
    }

    // Build post 
    post, err := pb.Build(renderHashtags, languages)
    if err != nil {
        return "", "", err
    }

    return c.CreateRecordPost(ctx, post)
}


// Create a simple post with text
func NewPostBuilder(text string) PostBuilder {
	return PostBuilder{
		Text:  text,
		Facet: []Facet{},
	}
}

// Create a Richtext Post with facets
func (pb PostBuilder) WithFacet(ftype Facet_Type, value string, text string) PostBuilder {

	pb.Facet = append(pb.Facet, Facet{
		Ftype:   ftype,
		Value:   value,
		T_facet: text,
	})

	return pb
}

// Create a Post with external links
func (pb PostBuilder) WithExternalLink(title string, link url.URL, description string, thumb lexutil.LexBlob) PostBuilder {

	pb.Embed.Link.Title = title
	pb.Embed.Link.Uri = link
	pb.Embed.Link.Description = description
	pb.Embed.Link.Thumb = thumb

	return pb
}

// Create a Post with images
func (pb PostBuilder) WithImages(blobs []lexutil.LexBlob, images []Image) PostBuilder {

	pb.Embed.Images = images
	pb.Embed.UploadedImages = blobs

	return pb
}

// Build the request
func (pb PostBuilder) Build(renderHashtags bool, languages []string) (bsky.FeedPost, error) {

    post := bsky.FeedPost{ Langs: languages }

	post.Text = pb.Text
	post.LexiconTypeID = "app.bsky.feed.post"
	post.CreatedAt = time.Now().Format(time.RFC3339)

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
    if renderHashtags {
        hashtagRegex:= `(?:^|\s)(#[^\d\s]\S*)`
        matches := findRegexMatches(post.Text, hashtagRegex)
        for _, m := range matches {
            facet := &bsky.RichtextFacet{}
            features := []*bsky.RichtextFacet_Features_Elem{}
            feature := &bsky.RichtextFacet_Features_Elem{}
            
            feature = &bsky.RichtextFacet_Features_Elem{
                RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
                    LexiconTypeID: Facet_Tag.String(),
                    Tag: stripHashtag(m.Value),
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

	}

	// avoid error when trying to marshal empty field (*bsky.FeedPost_Embed)
	if len(pb.Embed.Images) != 0 || pb.Embed.Link.Title != "" {
		post.Embed = &FeedPost_Embed
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

// This function has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/gobot.go
// License: Apache 2.0
func (c *Client) UploadImage(ctx context.Context, image Image) (*lexutil.LexBlob, error) {

	getImage, err := getImageAsBuffer(image.Uri.String())
	if err != nil {
		log.Printf("Couldn't retrive the image: %v , %v", image, err)
	}

	resp, err := atproto.RepoUploadBlob(ctx, c.xrpcClient, bytes.NewReader(getImage))
	if err != nil {
		return nil, err
	}

	blob := lexutil.LexBlob{
		Ref:      resp.Blob.Ref,
		MimeType: resp.Blob.MimeType,
		Size:     resp.Blob.Size,
	}

	return &blob, nil
}

// This function has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/gobot.go
// License: Apache 2.0
func (c *Client) UploadImages(ctx context.Context, images []Image) ([]lexutil.LexBlob, error) {

    var blobs []lexutil.LexBlob

	for _, img := range images {
		getImage, err := getImageAsBuffer(img.Uri.String())
		if err != nil {
			log.Printf("Couldn't retrive the image: %v , %v", img, err)
		}

		resp, err := atproto.RepoUploadBlob(ctx, c.xrpcClient, bytes.NewReader(getImage))
		if err != nil {
			return nil, err
		}

		blobs = append(blobs, lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: resp.Blob.MimeType,
			Size:     resp.Blob.Size,
		})
	}
	return blobs, nil
}


// This function has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/gobot.go
// License: Apache 2.0
// Post to social app
func (c *Client) CreateRecordPost(ctx context.Context, post bsky.FeedPost) (string, string, error) {

	post_input := &atproto.RepoCreateRecord_Input{
		// collection: The NSID of the record collection.
		Collection: "app.bsky.feed.post",
		// repo: The handle or DID of the repo (aka, current account).
		Repo: c.xrpcClient.Auth.Did,
		// record: The record itself. Must contain a $type field.
		Record: &lexutil.LexiconTypeDecoder{Val: &post},
	}

	response, err := atproto.RepoCreateRecord(ctx, c.xrpcClient, post_input)
	if err != nil {
		return "", "", fmt.Errorf("unable to post, %v", err)
	}

	return response.Cid, response.Uri, nil
}
