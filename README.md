# Botsky

A Bluesky API client in Go with useful features for writing automated bots.

## Design

This client does not simply wrap the various API calls (see [api docs](https://docs.bsky.app/docs/api/com-atproto-repo-get-record)), as this is already implemented by github.com/bluesky-social/indigo/xrpc. Instead, it provides useful abstractions for working with the API and the ecosystem in general. Note that "useful" is currently defined by me and the library is thus opinionated, but feel free to open issues to discuss requests regarding the design.

decentralization/federation
=> when possible, interact with atproto instead of Bluesky API

## Features

- Session management, authentication, auto-refresh
- Posts:
  - get from pds/collection
  - get by uri
  - get rich posts & post views from bsky appview including like counts etc.
  - create posts with links, mentions, tags, images
  - reply to and quote posts
  - repost
  - delete own posts

## TODO

- get repo contents

- error handling: whenever we return an error, also add a prefix where/why it happened
- integrate jetstream, set up event listeners

  - => could this be implemented via the API app.bsky.notification -> registerPush/listNotifications?

  - listen for mentions
  - listen for replies
  - other listeners

- builtin adjustable rate limiting

- social graph, user profiles, followers

- further api integration (lists, feeds, graph, labels, etc.)

- refer to Bluesky guidelines related to API, bots, etc., bots should adhere to guidelines

- trust, verification, cryptography: in general the server hosting the PDS is not trusted, should verify data returned by it

- reliance on Bluesky's (the company) AppView... can we make it atproto-native? include smth like a trustBluesky flag?

## Acknowledgements

This library is partially inspired by and adapted from

- Bluesky Go Bot framework: https://github.com/danrusei/gobot-bsky
- Go-Bluesky client library: https://github.com/karalabe/go-bluesky
