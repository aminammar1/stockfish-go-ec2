package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func landingPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>Stockfish EC2 Service</title>
    <style>
      :root { --bg:#0b1020; --card:#121a33; --text:#e6e8ef; --muted:#a9b0c3; --link:#8ab4ff; --ok:#62d26f; }
      body { margin:0; font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Ubuntu, Cantarell, Noto Sans, Arial; background: radial-gradient(1200px 600px at 20% 10%, #19224a 0%, var(--bg) 55%); color: var(--text); }
      .wrap { max-width: 920px; margin: 0 auto; padding: 40px 20px; }
      .card { background: color-mix(in oklab, var(--card) 92%, black); border: 1px solid rgba(255,255,255,0.08); border-radius: 16px; padding: 22px; box-shadow: 0 10px 30px rgba(0,0,0,0.35); }
      h1 { margin: 0 0 6px; font-size: 28px; }
      .sub { color: var(--muted); margin: 0 0 18px; }
      a { color: var(--link); text-decoration: none; }
      a:hover { text-decoration: underline; }
      .grid { display: grid; grid-template-columns: 1fr; gap: 14px; margin-top: 14px; }
      @media (min-width: 760px) { .grid { grid-template-columns: 1fr 1fr; } }
      .pill { display:inline-block; padding: 6px 10px; border-radius: 999px; background: rgba(98,210,111,0.12); color: var(--ok); border: 1px solid rgba(98,210,111,0.22); font-size: 12px; }
      pre { margin: 10px 0 0; padding: 12px; border-radius: 12px; background: rgba(0,0,0,0.25); overflow:auto; border: 1px solid rgba(255,255,255,0.08); }
      code { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
      .k { color: #ffd6a7; }
    </style>
  </head>
  <body>
    <div class="wrap">
      <div class="card">
        <div class="pill">online</div>
        <h1>Stockfish EC2 Service</h1>
        <p class="sub">A tiny API that proxies Stockfish over SSH.</p>

        <div class="grid">
          <div class="card">
            <h2 style="margin:0 0 8px; font-size:18px">Docs</h2>
            <div><a href="/swagger/index.html">Swagger UI</a></div>
            <div><a href="/swagger/doc.json">OpenAPI JSON</a></div>
          </div>

          <div class="card">
            <h2 style="margin:0 0 8px; font-size:18px">Try It</h2>
            <div><span class="k">GET</span> <a href="/api/v1/health">/api/v1/health</a></div>
            <div><span class="k">POST</span> /api/v1/analyze</div>
            <pre><code>curl -sS -X POST \
  -H 'Content-Type: application/json' \
  -d '{"fen":"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1","depth":12}' \
  /api/v1/analyze</code></pre>
          </div>
        </div>
      </div>
    </div>
  </body>
</html>`))
	}
}
