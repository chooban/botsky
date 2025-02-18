package botsky

import (
	"context"
	"fmt"
	"time"

	"github.com/davhofer/indigo/api/atproto"
	"github.com/davhofer/indigo/xrpc"
	"github.com/golang-jwt/jwt/v5"
)

// Extracts the remaining time until expiry from a jwt string
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

// Update the clients auth info with the given JWTs, handle, and did.
//
// This will also start a new goroutine to refresh the session before this one expires
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

// Refreshes the client's session once the timer expires
func (c *Client) RefreshSession(ctx context.Context, timer *time.Timer) {
	// wait until timer fires
	<-timer.C

	c.refreshProcessLock.Lock()
	defer c.refreshProcessLock.Unlock()

	// check that RefreshJWT is still (for some time) valid
	auth := c.xrpcClient.GetAuthAsync()
	if tRemaining, _ := getJwtTimeRemaining(auth.RefreshJwt); tRemaining > 30*time.Second {
		// set access jwt to refresh jwt
		auth.AccessJwt = auth.RefreshJwt
		c.xrpcClient.SetAuthAsync(auth)
		// refresh the session
		session, err := atproto.ServerRefreshSession(ctx, c.xrpcClient)

		if err != nil { // log error if it happened
			logger.Println("RefreshSession error (ServerRefreshSession):", err)
		} else if err := c.UpdateAuth(ctx, session.AccessJwt, session.RefreshJwt, session.Handle, session.Did); err != nil { // otherwise try to update auth
			logger.Println("RefreshSession error (UpdateAuth):", err)
		} else {
			// if neither of the above returned an error, we successfully updated auth
			return
		}
	}

	// otherwise, perform full auth
	if err := c.Authenticate(ctx); err != nil {
		logger.Println("RefreshSession error (Authenticate):", err)
	}
}

// Authenticates the client with the given credentials and updates its auth info.
//
// A background goroutine to automatically refresh the session is started through client.UpdateAuth
func (c *Client) Authenticate(ctx context.Context) error {
	// reset auth
	c.xrpcClient.SetAuthAsync(xrpc.AuthInfo{})
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
