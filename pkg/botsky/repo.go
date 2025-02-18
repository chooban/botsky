package botsky

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/davhofer/indigo/api/atproto"
	"github.com/davhofer/indigo/api/bsky"
	lexutil "github.com/davhofer/indigo/lex/util"
	util "github.com/davhofer/indigo/util"
)

// TODO: download image function from embed/repo, using SyncGetBlob
// useful to extract images e.g. from posts

// TODO: info about user/profile

// TODO: functions to get likes, follows, followers, posts, etc.
// use curser to go through all pages

// Get all collections available on the repo.
func (c *Client) RepoGetCollections(ctx context.Context, handleOrDid string) ([]string, error) {
	output, err := atproto.RepoDescribeRepo(ctx, c.xrpcClient, handleOrDid)
	if err != nil {
		return nil, fmt.Errorf("RepoGetCollections error (RepoDescribeRepo): %v", err)
	}
	return output.Collections, nil
}

// Get all recors of the specified collection from the given repo.
func (c *Client) RepoGetRecords(ctx context.Context, handleOrDid string, collection string, limit int) ([]*atproto.RepoListRecords_Record, error) {

	var records []*atproto.RepoListRecords_Record
	cursor, lastCid := "", ""

	// iterate until we got all records
	for {
		// query repo for collection with updated cursor
		output, err := atproto.RepoListRecords(ctx, c.xrpcClient, collection, cursor, 100, handleOrDid, false, "", "")
		if err != nil {
			return nil, fmt.Errorf("RepoGetRecords error (RepoListRecords): %v", err)
		}

		// stop if no records returned
		// or we get one we've already seen (maybe a repo doesn't support cursor?)
		if len(output.Records) == 0 || lastCid == output.Records[len(output.Records)-1].Cid {
			break
		}
		lastCid = output.Records[len(output.Records)-1].Cid
		// store all record uris
		records = append(records, output.Records...)

		// if we have more records than the requested limit, stop
		// limit -1 indicates no upper limit, i.e. get all record
		if limit != -1 && len(records) >= limit {
			break
		}
		// update cursor
		cursor = *output.Cursor
	}

	// don't return more than the requested limit
	var end int
	if limit == -1 {
		end = len(records)
	} else {
		end = min(len(records), limit)
	}
	return records[:end], nil
}

// Get record uris from the repo in the given collection.
//
// Set limit = -1 in order to get all records.
func (c *Client) RepoGetRecordUris(ctx context.Context, handleOrDid string, collection string, limit int) ([]string, error) {
	records, err := c.RepoGetRecords(ctx, handleOrDid, collection, limit)
	if err != nil {
		return nil, fmt.Errorf("RepoGetRecordUris error (RepoGetRecords): %v", err)
	}
	uris := make([]string, len(records))
	for i, r := range records {
		uris[i] = r.Uri
	}
	return uris, nil
}

// Get the record at the specified uri and decode as the given result type (type of *resultPointer).
// Result is stored in the object referenced by resultPointer.
//
// E.g. var post bsky.FeedPost;
// RepoGetRecordAsType(ctx, postUri, &feedPost)
func (c *Client) RepoGetRecordAsType(ctx context.Context, recordUri string, resultPointer cborUnmarshaler) error {
	parsedUri, err := util.ParseAtUri(recordUri)
	if err != nil {
		return fmt.Errorf("RepoGetCidOfRecord error (ParseAtUri): %v", err)
	}
	record, err := atproto.RepoGetRecord(ctx, c.xrpcClient, "", parsedUri.Collection, parsedUri.Did, parsedUri.Rkey)
	if err != nil {
		return fmt.Errorf("RepoGetRecordAsType error (RepoGetRecord): %v", err)
	}
	return decodeRecordAsLexicon(record.Value, resultPointer)

}

// Get the FeedPost struct and the post CID given its Uri.
func (c *Client) RepoGetPostAndCid(ctx context.Context, postUri string) (bsky.FeedPost, string, error) {
	var post bsky.FeedPost
	parsedUri, err := util.ParseAtUri(postUri)
	if err != nil {
		return post, "", fmt.Errorf("RepoGetPostAndCid error (ParseAtUri): %v", err)
	}
	record, err := atproto.RepoGetRecord(ctx, c.xrpcClient, "", parsedUri.Collection, parsedUri.Did, parsedUri.Rkey)
	if err != nil {
		return post, "", fmt.Errorf("RepoGetPostAndCid error (RepoGetRecord): %v", err)
	}
	if err := decodeRecordAsLexicon(record.Value, &post); err != nil {
		return post, "", fmt.Errorf("RepoGetPostAndCid error (DecodeRecordAsLexicon): %v", err)
	}
	return post, *record.Cid, nil
}

// Delete the bots post with the given uri.
func (c *Client) RepoDeletePost(ctx context.Context, postUri string) error {
	parsedUri, err := util.ParseAtUri(postUri)
	if err != nil {
		return fmt.Errorf("RepoDeletePost error (ParseAtUri): %v", err)
	}
	_, err = atproto.RepoDeleteRecord(ctx, c.xrpcClient, &atproto.RepoDeleteRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       c.Handle,
		Rkey:       parsedUri.Rkey,
	})
	if err != nil {
		return fmt.Errorf("RepoDeletePost error (RepoDeleteRecord): %v", err)
	}
	return nil
}

// Delete all posts in the bots repository.
func (c *Client) RepoDeleteAllPosts(ctx context.Context) error {
	postUris, err := c.RepoGetRecordUris(ctx, c.Handle, "app.bsky.feed.post", -1)
	if err != nil {
		return fmt.Errorf("RepoDeleteAllPosts error (RepoGetRecordUris): %v", err)
	}
	logger.Println("Deleting", len(postUris), "posts from repo")

	for _, uri := range postUris {
		err = c.RepoDeletePost(ctx, uri)
		if err != nil {
			return fmt.Errorf("RepoDeleteAllPosts error (RepoDeletePost): %v", err)
		}
	}
	return nil
}

// Upload a single image to the repo.
//
// This function has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/gobot.go
// License: Apache 2.0
func (c *Client) RepoUploadImage(ctx context.Context, image imageSourceParsed) (*lexutil.LexBlob, error) {

	getImage, err := getImageAsBuffer(image.Uri.String())
	if err != nil {
		log.Printf("Couldn't retrive the image: %v , %v", image, err)
	}

	resp, err := atproto.RepoUploadBlob(ctx, c.xrpcClient, bytes.NewReader(getImage))
	if err != nil {
		return nil, fmt.Errorf("RepoUploadImage error (RepoUploadBlob): %v", err)
	}

	blob := lexutil.LexBlob{
		Ref:      resp.Blob.Ref,
		MimeType: resp.Blob.MimeType,
		Size:     resp.Blob.Size,
	}

	return &blob, nil
}

// Upload the provided images to the repo.
//
// This function has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/gobot.go
// License: Apache 2.0
func (c *Client) RepoUploadImages(ctx context.Context, images []imageSourceParsed) ([]lexutil.LexBlob, error) {

	blobs := make([]lexutil.LexBlob, 0, len(images))

	for _, img := range images {
		getImage, err := getImageAsBuffer(img.Uri.String())
		if err != nil {
			log.Printf("Couldn't retrive the image: %v , %v", img, err)
		}

		resp, err := atproto.RepoUploadBlob(ctx, c.xrpcClient, bytes.NewReader(getImage))
		if err != nil {
			return nil, fmt.Errorf("RepoUploadImages error (RepoUploadBlob): %v", err)
		}

		blobs = append(blobs, lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: resp.Blob.MimeType,
			Size:     resp.Blob.Size,
		})
	}
	return blobs, nil
}

// Create new post FeedPost record in the given repo.
//
// This function has been modified from its original version.
// Original source: https://github.com/danrusei/gobot-bsky/blob/main/gobot.go
// License: Apache 2.0
// Post to social app
func (c *Client) RepoCreatePostRecord(ctx context.Context, post bsky.FeedPost) (string, string, error) {

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
