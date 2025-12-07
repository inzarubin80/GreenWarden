API: author_boosty_url and author_avatar_url in violation.requests[]

Summary
-------
Added `author_boosty_url` and `author_avatar_url` fields to each item in `requests[]` returned by `GET /api/violations/{id}`.
These fields contain, respectively, the current `boosty_url` and `avatar_url` of the user who created the request (or `null`/omitted if none).

Behavior
--------
- Values are resolved at response time by joining `violation_requests.created_by_user_id` to `users.user_id` and taking `users.boosty_url` / `users.avatar_url`.
- The values are not denormalized in `violation_requests` table (we did not add persistent columns). Therefore changes to a user's profile are visible immediately on the next `GET /api/violations/{id}` call.
- For single-request endpoints (e.g. `GET /api/violations/{violation_id}/requests/{request_id}`) the handler also returns both fields.

Response contract (snippet)
---------------------------
Within the existing `GET /api/violations/{id}` response, each `requests[]` object now includes:

```json
"requests": [
  {
    "id": "req-uuid",
    "status": "open",
    "created_by_user_id": 123,
    "comment": "text",
    "created_at": "2025-12-08T12:34:56Z",
    "photos": [ /* ... */ ],
    "likes": 5,
    "dislikes": 0,
    "user_vote": "like",
    "author_boosty_url": "https://boosty.to/username",  // or null / absent
    "author_avatar_url": "https://cdn.example.com/avatars/42/....jpg" // or null / absent
  }
]
```

Notes for mobile/web clients
---------------------------
- Treat `author_boosty_url` as optional: it may be absent or null if the author didn't set a Boosty link.
- If present, show a \"Support\" / \"Поддержать\" button linked to this URL.
- Treat `author_avatar_url` as optional: if absent/null, show a fallback avatar (e.g. initials); if present, load avatar image from this URL.
- Because the value is live, no special cache invalidation is required on the backend side.

Manual verification steps
-------------------------
1. Ensure some user has `boosty_url` and `avatar_url` set in `users` table. Example SQL:
   ```sql
   UPDATE users SET boosty_url = 'https://boosty.to/testuser', avatar_url = 'https://cdn.example.com/avatars/1/test.jpg' WHERE user_id = 1;
   ```
2. Create a violation and a request as that user (or find existing IDs).
3. Call `GET /api/violations/{violation_id}` and verify:
   - `requests[i].author_boosty_url` equals `'https://boosty.to/testuser'`.
   - `requests[i].author_avatar_url` equals `'https://cdn.example.com/avatars/1/test.jpg'`.
4. Change the user's `boosty_url` and/or `avatar_url` and call GET again — verify new values are returned.

Implementation notes
--------------------
- SQL changed: `GetViolationRequestsByViolationID` now LEFT JOINs `users` and selects `u.boosty_url AS author_boosty_url`.
- `model.ViolationRequest` extended with `AuthorBoostyURL string json:"author_boosty_url,omitempty"`.
- Repository and handlers were updated to map this value into the response DTO.


