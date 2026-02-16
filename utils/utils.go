package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// DoPost sends a POST request (fire and forget).
func DoPost(url string, body string) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "text/ping")
	http.DefaultClient.Do(req)
}

// DoRequest performs an HTTP request (GET or POST).
func DoRequest(
	method string,
	pageURL string,
	body []byte,
	contentType string,
	formData url.Values,
	referrerPolicy string,
	fromURL string,
) (*http.Response, error) {
	var httpReq *http.Request
	var err error

	if method == "POST" {
		if body != nil && contentType != "" {
			// Multipart form data (file upload)
			httpReq, err = http.NewRequest("POST", pageURL, bytes.NewReader(body))
			if err != nil {
				return nil, err
			}
			httpReq.Header.Set("Content-Type", contentType)
		} else {
			httpReq, err = http.NewRequest("POST", pageURL, strings.NewReader(formData.Encode()))
			if err != nil {
				return nil, err
			}
			httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		httpReq, err = http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return nil, err
		}
	}

	parsed, err := url.Parse(fromURL)
	if err == nil {
		fmt.Println(parsed)
		originURL := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)

		switch referrerPolicy {
		case "origin":
			httpReq.Header.Set("Referer", originURL)
		case "same-origin":
			if strings.EqualFold(parsed.Host, httpReq.URL.Host) {
				httpReq.Header.Set("Referer", fromURL)
			}
		case "strict-origin-when-cross-origin":
			if strings.EqualFold(parsed.Host, httpReq.URL.Host) {
				httpReq.Header.Set("Referer", fromURL)
			} else {
				httpReq.Header.Set("Referer", originURL)
			}
		case "unsafe-url":
			httpReq.Header.Set("Referer", fromURL)
		default:
		}
	}

	return http.DefaultClient.Do(httpReq)
}

// ParseHTMLSizeAttribute parses width/height attributes.
// Supports percent (relative to containerWidth), px, and raw numbers.
func ParseHTMLSizeAttribute(value string, containerWidth float64) float64 {
	v := strings.TrimSpace(value)
	if strings.HasSuffix(v, "%") {
		num := strings.TrimSuffix(v, "%")
		if pct, err := strconv.ParseFloat(num, 64); err == nil && pct >= 0 {
			return containerWidth * pct / 100.0
		}
		return 0
	}

	lower := strings.ToLower(v)
	if strings.HasSuffix(lower, "px") {
		v = strings.TrimSuffix(lower, "px")
	}
	if n, err := strconv.ParseFloat(v, 64); err == nil && n > 0 {
		return n
	}

	return 0
}
