package botsky

import (
	"context"
	"fmt"
	"time"

	"github.com/davhofer/indigo/api/atproto"
	"github.com/davhofer/indigo/xrpc"
	"github.com/golang-jwt/jwt/v5"
)

func getJwtTimeRemaining(tokenString string) (time.Duration, error) {
	// TODO: improve this?
	token, _, _ := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{}) 

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("Cannot parse claims from JWT")
	}
	expTime, err := claims.GetExpirationTime()
	if err != nil {
        return 0, fmt.Errorf("Error when trying to get JWT expiration time")
	}
	// Calculate the time remaining
	expTimeUnix := time.Unix(expTime.Unix(), 0)
	return time.Until(expTimeUnix), nil
}

func (c *Client) UpdateAuth(ctx context.Context, accessJwt string, refreshJwt string, handle string, did string) error {
	c.xrpcClient.SetAuthAsync(xrpc.AuthInfo{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
		Handle:     handle,
		Did:        did,
	})

    if c.chatClient != nil {
        c.chatClient.SetAuthAsync(xrpc.AuthInfo{
            AccessJwt:  accessJwt,
            RefreshJwt: refreshJwt,
            Handle:     handle,
            Did:        did,
        })
    }

	// Start timer for expiration of AccessJWT and session refresh
	// parse time until expiration from accessJwt
	tRemaining, err := getJwtTimeRemaining(accessJwt)
	if err != nil {
        return fmt.Errorf("UpdateAuth error: %v", err)
	}

	timer := time.NewTimer(tRemaining - time.Minute)
	// start refresher goroutine in background
	go c.RefreshSession(ctx, timer)
    return nil
}

func (c *Client) RefreshSession(ctx context.Context, timer *time.Timer) {
	// wait until timer fires
	<-timer.C

	c.refreshProcessLock.Lock()
	defer c.refreshProcessLock.Unlock()

	// check that RefreshJWT is still (for some time) valid
	if tRemaining, _ := getJwtTimeRemaining(c.xrpcClient.Auth.RefreshJwt); tRemaining > 30*time.Second {
		// refresh the session
		session, err := atproto.ServerRefreshSession(ctx, c.xrpcClient)
		if err != nil {
			// TODO: how to handle this error?
            logger.Println("RefreshSession error (ServerRefreshSession):", err)
		}
        if err := c.UpdateAuth(ctx, session.AccessJwt, session.RefreshJwt, session.Handle, session.Did); err != nil {
            logger.Println("RefreshSession error (UpdateAuth):", err)
        }
	}

	// otherwise, perform full auth
	if err := c.Authenticate(ctx); err != nil {
        logger.Println("RefreshSession error (Authenticate):", err)
    }
}

// Authenticates the client with the given credentials
// Re-authenticating: First try to send RefreshJwt, if that fails create fully new session
func (c *Client) Authenticate(ctx context.Context) error {
	// create new session and authenticate with handle and appkey
	sessionCredentials := &atproto.ServerCreateSession_Input{
		Identifier: c.Handle,
		Password:   c.appkey,
	}
	session, err := atproto.ServerCreateSession(ctx, c.xrpcClient, sessionCredentials)
	if err != nil {
		// TODO: how to handle this error? if called as goroutine, it will be lost
		return fmt.Errorf("Authenticate error (ServerCreateSession): %v", err)
	}
    if err := c.UpdateAuth(ctx, session.AccessJwt, session.RefreshJwt, session.Handle, session.Did); err != nil {
		return fmt.Errorf("Authenticate error (UpdateAuth): %v", err)
    }
	return nil
}
