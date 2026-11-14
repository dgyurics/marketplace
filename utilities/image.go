package utilities

import (
	"encoding/base64"
	"net/http"
)

// DataURL returns the bytes as a base64 data URL, such as
// "data:image/png;base64,...", so they can be sent inline in a request.
// Drawback: can be large, and comes with an overhead of about 33% due to base64 encoding
func DataURL(data []byte) string {
	return "data:" + http.DetectContentType(data) + ";base64," + base64.StdEncoding.EncodeToString(data)
}
