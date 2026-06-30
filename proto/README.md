# proto/

The **contracts**. Every service's gRPC interface lives here as a `.proto` file.

> ⚠️ SHARED FILE ZONE. Changing anything here affects both developers.
> Agree on a proto change together before committing it (CLAUDE.md / roadmap rule 1).

Planned files (write them together in Phase 0, step 0.2):

```
auth.proto     — Register, VerifyOTP, Login, RefreshToken
user.proto     — GetProfile, UpdateProfile, Follow, Unfollow, GetFollowers
posts.proto    — CreatePost, DeletePost, GetPost, LikePost, AddComment
feed.proto     — GetFeed, GetExplore
notif.proto    — GetNotifications, MarkAsRead
search.proto   — Search
```

Event schemas (async contracts) are documented in `docs/04-communication.md`:
`post.created`, `post.liked`, `comment.created`, `user.followed`, `user.updated`.
