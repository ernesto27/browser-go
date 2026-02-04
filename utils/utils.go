package utils

import (
	"net/http"
	"strings"
)

// PostAsync sends a POST request asynchronously (fire and forget).
func DoPost(url string, body string) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "text/ping")
	http.DefaultClient.Do(req)

}
