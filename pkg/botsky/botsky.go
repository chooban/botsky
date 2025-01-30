package botsky

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
    "bufio"
    "syscall"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"

    "golang.org/x/term"
)

const DefaultServer = "https://bsky.social"

func GetEnvCredentials() (string, string, error) {
	handle := os.Getenv("BOTSKY_HANDLE")
	appkey := os.Getenv("BOTSKY_APPKEY")
    if handle == "" || appkey == "" {
        return "", "", fmt.Errorf("BOTSKY_HANDLE or BOTSKY_APPKEY env variable not set")
    }
	return handle, appkey, nil
}

func GetCLICredentials() (string, string, error) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter account handle: ")
    handle, err := reader.ReadString('\n')
    if err != nil {
        return "", "", err
    }

    fmt.Print("Enter appkey: ")
    byteAppkey, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return "", "", err
    }

    appkey := string(byteAppkey)
    return strings.TrimSpace(handle), strings.TrimSpace(appkey), nil
}

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

func (c *Client) CanGetPreferences(ctx context.Context) bool {
    _, err := bsky.ActorGetPreferences(ctx, c.xrpcClient)
    return err != nil
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

