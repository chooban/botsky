# Botsky

A Bluesky API client in Go with useful features for writing automated bots.

## Design

This client does not simply wrap the various API calls (see [api docs](https://docs.bsky.app/docs/api/com-atproto-repo-get-record)), as this is already implemented by github.com/bluesky-social/indigo/xrpc. Instead, it provides useful abstractions for working with the API and the ecosystem in general. Note that "useful" is currently defined by me and the library is thus opinionated, but feel free to open issues to discuss requests regarding the design.

## Features

- Session management, authentication, auto-refresh
- Posts: create posts with links, mentions, tags, images

## TODO

- posts: reply, repost
- integrate jetstream, set up event listeners

  - => could this be implemented via the API app.bsky.notification -> registerPush/listNotifications?

  - listen for mentions
  - listen for replies
  - other listeners

- builtin adjustable rate limiting

- social graph, user profiles, followers

- further api integration (lists, feeds, graph, labels, etc.)

- refer to Bluesky guidelines related to API, bots, etc., bots should adhere to guidelines

## Acknowledgements

This library is partially inspired by and adapted from

- Bluesky Go Bot framework: https://github.com/danrusei/gobot-bsky
- Go-Bluesky client library: https://github.com/karalabe/go-bluesky
