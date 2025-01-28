package botsky

import (
	"context"
	"github.com/bluesky-social/indigo/xrpc"
	"net/http"
	"os"
	"sync"
)

const DefaultServer = "https://bsky.social"

type Client struct {
	xrpcClient *xrpc.Client
	handle     string
	appkey     string
	// make sure only one refresher runs at a time
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

func GetEnvCredentials() (string, string) {
	handle := os.Getenv("BOTSKY_HANDLE")
	appkey := os.Getenv("BOTSKY_APPKEY")
	return handle, appkey
}
