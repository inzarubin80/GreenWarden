# VK / OpenGraph preview for `/violations/:id`

## Why it’s needed
Social crawlers (VK, etc.) usually **don’t execute JS**, so SPA pages won’t produce a rich preview unless OG tags are served from **server-side HTML**.

We implemented a Go handler:
- `GET /violations/{id}` (on the Go server) returns an HTML page with `og:*` tags.

## Recommended routing (nginx)
You want **bots** to hit the Go server HTML, but **humans** to keep receiving the SPA.

Example nginx config snippet:

```nginx
# Detect bots (VK + common social crawlers)
map $http_user_agent $is_social_bot {
  default 0;
  ~*vkshare 1;
  ~*vkbot 1;
  ~*facebookexternalhit 1;
  ~*twitterbot 1;
  ~*telegrambot 1;
  ~*whatsapp 1;
  ~*okhttp 1;
}

# Upstream to Go server (adjust host/port)
upstream greenwarden_api {
  server 127.0.0.1:8090;
}

server {
  # ... existing server config ...

  # For share links: bots -> Go HTML with OG, humans -> SPA
  location ~ ^/violations/ {
    if ($is_social_bot) {
      proxy_set_header Host $host;
      proxy_set_header X-Forwarded-Proto $scheme;
      proxy_set_header X-Forwarded-Host $host;
      proxy_pass http://greenwarden_api;
      break;
    }

    # humans: serve SPA
    root /usr/share/nginx/html;
    try_files $uri $uri/ /index.html;
  }

  # ... rest (/, /assets, etc) ...
}
```

## Notes
- Set `PUBLIC_WEB_BASE=https://green-warden.ru` on the Go server so `og:url` is always canonical.
- `og:image` uses the first “problem photo” if present (CDN/presigned URL) so bots can fetch it.

