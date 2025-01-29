package botsky

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
)

func RkeyFromUri(uri string) (string, error) {
    splits := strings.Split(uri, "/")
    if len(splits) > 1 {
        return splits[len(splits)-1], nil
    }
    return "", fmt.Errorf("unable to split uri and get rkey.")
}



func (c *Client) RepoGetCollections(ctx context.Context, handleOrDid string) ([]string, error) {
	output, err := atproto.RepoDescribeRepo(ctx, c.xrpcClient, handleOrDid)
	if err != nil {
		return nil, err
	}
	return output.Collections, nil
}

// TODO: functions to get likes, follows, followers, posts, etc.
// use curser to go through all pages

func (c *Client) RepoGetAllRecordUris(ctx context.Context, handleOrDid string, collection string) ([]string, error) {

    var uris []string
    cursor, lastCid := "", ""
    
    // iterate until we got all records
    for {
        // query repo for collection with updated cursor
        output, err := atproto.RepoListRecords(ctx, c.xrpcClient, collection, cursor, 100, handleOrDid, false, "", "")
        if err != nil {
            return nil, err
        }

        // abort if no records returned or we get one we've already seen (maybe a repo 
        // doesn't support cursor?)
        if len(output.Records) == 0 || lastCid == output.Records[len(output.Records)-1].Cid {
            break
        }
        // store all record uris
        for _, record := range output.Records {
            uris = append(uris, record.Uri)
        }
        // update cursor
        cursor = *output.Cursor
    }
    return uris, nil
}

// TODO: improve Post class, include post text and other data...
func (c *Client) GetPosts(ctx context.Context, handleOrDid string) ([]*bsky.FeedDefs_PostView, error) {
    // get all post uris
    postUris, err := c.RepoGetAllRecordUris(ctx, handleOrDid, "app.bsky.feed.post")
    if err != nil {
        return nil, err
    }

    // hydrate'em
    var postViews []*bsky.FeedDefs_PostView
    for i := 0; i < len(postUris); i+=25 {
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

func (c *Client) DeletePost(ctx context.Context, postUri string) error {
    rkey, err := RkeyFromUri(postUri)
    if err != nil {
        return err 
    }
    _, err = atproto.RepoDeleteRecord(ctx, c.xrpcClient, &atproto.RepoDeleteRecord_Input{
        Collection: "app.bsky.feed.post",
        Repo: c.handle, 
        Rkey: rkey,
        })
    if err != nil {
        return err 
    }
    return nil
}

func (c *Client) DeleteAllPosts(ctx context.Context) error {
    posts, err := c.GetPosts(ctx, c.handle)
    if err != nil {
        return err
    }
    fmt.Println("Deleting", len(posts), "posts")

    for _, post := range posts {
        err = c.DeletePost(ctx, post.Uri)
        if err != nil {
            return err
        }
    }
    return nil
}








func (c *Client) RepoListRecords(ctx context.Context, handleOrDid string, collection string, limit int) error {
	var limit64 int64
	switch {
	case limit > 100:
		fmt.Println("Limit was set to maxVal 100")
		limit64 = 100
	case limit < 1:
		fmt.Println("Limit was set to minVal 1")
		limit64 = 1
	default:
		limit64 = int64(limit)
	}
	output, err := atproto.RepoListRecords(ctx, c.xrpcClient, collection, "", limit64, handleOrDid, false, "", "")
	if err != nil {
		return err
	}
    fmt.Println(output.Records)
	fmt.Println("\nListing records:")
	for _, record := range output.Records {
		fmt.Println(record.Uri, record.Value.Val)
        
	}
	fmt.Println()
	return nil
}
