package main

import (
	"botsky/pkg/botsky"
	"context"
    "fmt"
    "time"
)

func main() {
    ctx := context.Background()

    defer fmt.Println("botsky is going to bed...")

    handle, appkey := botsky.GetEnvCredentials()

    if handle == "" || appkey == "" {
        fmt.Println("Handle and AppKey for authentication need to be accessible as environment variables BOTSKY_HANDLE and BOTSKY_APPKEY")
        return
    }

    client, err := botsky.NewClient(ctx, botsky.DefaultServer, handle, appkey)
    if err != nil {
        fmt.Println(err)
        return
    }

    err = client.Authenticate(ctx)
    if err != nil {
        return
    }

    fmt.Println("Authentication successful")

    time.Sleep(5 * time.Second)


}
