package demo

import "embed"

//go:embed templates/index.html static/app.js static/styles.css
var assets embed.FS
