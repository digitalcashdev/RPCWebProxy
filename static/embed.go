package static

import "embed"

// wget https://andybrewer.github.io/mvp/mvp.css

//go:embed index.html mvp.css public-rpcs.json
var FS embed.FS
var Prefix = ""
