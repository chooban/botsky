package botsky

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/golang-jwt/jwt/v5"
)

func getJwtTimeRemaining(tokenString string) (time.Duration, error) {
	// TODO: improve this?
	token, _, _ := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{}) // new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Err: cannot parse claims")
		return 0, fmt.Errorf("cannot parse claims from jwt")
	}
	expTime, err := claims.GetExpirationTime()
	if err != nil {
		fmt.Println(err)
	}
	// Calculate the time remaining
	expTimeUnix := time.Unix(expTime.Unix(), 0)
	return time.Until(expTimeUnix), nil
}

func (c *Client) UpdateAuth(ctx context.Context, accessJwt string, refreshJwt string, handle string, did string) {
	c.xrpcClient.Auth = &xrpc.AuthInfo{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
		Handle:     handle,
		Did:        did,
	}

	fmt.Println("- Update Auth -\nhandle:", handle, "\ndid:", did)

	// Start timer for expiration of AccessJWT and session refresh
	// parse time until expiration from accessJwt
	tRemaining, err := getJwtTimeRemaining(accessJwt)
	if err != nil {
		// TODO: how to handle this error
	}

	fmt.Println("AccessJwt time remaining:", tRemaining)

	timer := time.NewTimer(tRemaining - 30*time.Second)
	// start refresher goroutine in background
	go c.RefreshSession(ctx, timer)
}

func (c *Client) RefreshSession(ctx context.Context, timer *time.Timer) error {
	id := rand.IntN(1000000)
	fmt.Println("id", id, "| spawned new RefreshSession routine")
	// wait until timer fires
	<-timer.C

	fmt.Println("id", id, "| timer fired, waiting for mutex")

	c.refreshMutex.Lock()
	defer c.refreshMutex.Unlock()

	fmt.Println("id", id, "| acquired mutex")
	defer fmt.Println("id", id, "| [returning...]")

	// check that RefreshJWT is still (for some time) valid
	if tRemaining, _ := getJwtTimeRemaining(c.xrpcClient.Auth.RefreshJwt); tRemaining > 30*time.Second {
		// refresh the session
		session, err := atproto.ServerRefreshSession(ctx, c.xrpcClient)
		if err != nil {
			// TODO: how to handle this error?
			err = fmt.Errorf("Session refresh failed: %v", err)
			fmt.Println(err)
			return err
		}
		c.UpdateAuth(ctx, session.AccessJwt, session.RefreshJwt, session.Handle, session.Did)
		return nil
	}

	// otherwise, perform full auth
	err := c.Authenticate(ctx)

	if err != nil {
		// TODO: how to handle this error?
		err = fmt.Errorf("Session authentication failed: %v", err)
		fmt.Println(err)
	}
	return err
}

// Authenticates the client with the given credentials
// Re-authenticating: First try to send RefreshJwt, if that fails create fully new session
func (c *Client) Authenticate(ctx context.Context) error {
	// create new session and authenticate with handle and appkey
	sessionCredentials := &atproto.ServerCreateSession_Input{
		Identifier: c.handle,
		Password:   c.appkey,
	}
	session, err := atproto.ServerCreateSession(ctx, c.xrpcClient, sessionCredentials)
	if err != nil {
		// TODO: how to handle this error? if called as goroutine, it will be lost
		err = fmt.Errorf("Session creation failed: %v", err)
		fmt.Println(err)
		return err
	}
	c.UpdateAuth(ctx, session.AccessJwt, session.RefreshJwt, session.Handle, session.Did)
	return nil
}
