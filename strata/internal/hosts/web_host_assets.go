package hosts

import _ "embed"

//go:embed web/dist/index.html
var webHostIndexHTML []byte
