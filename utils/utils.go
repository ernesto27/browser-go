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

// HTTPRequest holds the parameters for DoRequest.
type HTTPRequest struct {
	Method         string
	URL            string
	Body           []byte
	ContentType    string
	FormData       url.Values
	ReferrerPolicy string
	FromURL        string
}

// DoRequest performs an HTTP request (GET or POST).
func DoRequest(req HTTPRequest) (*http.Response, error) {
	method := req.Method
	pageURL := req.URL
	body := req.Body
	contentType := req.ContentType
	formData := req.FormData
	referrerPolicy := req.ReferrerPolicy
	fromURL := req.FromURL

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

		isDowngrade := parsed.Scheme == "https" && httpReq.URL.Scheme == "http"
		isSameOrigin := strings.EqualFold(parsed.Host, httpReq.URL.Host)

		switch referrerPolicy {
		case "no-referrer":
			// Never send Referer
		case "no-referrer-when-downgrade":
			if !isDowngrade {
				httpReq.Header.Set("Referer", fromURL)
			}
		case "origin":
			httpReq.Header.Set("Referer", originURL)
		case "origin-when-cross-origin":
			if isSameOrigin {
				httpReq.Header.Set("Referer", fromURL)
			} else {
				httpReq.Header.Set("Referer", originURL)
			}
		case "same-origin":
			if isSameOrigin {
				httpReq.Header.Set("Referer", fromURL)
			}
		case "strict-origin":
			if !isDowngrade {
				httpReq.Header.Set("Referer", originURL)
			}
		case "strict-origin-when-cross-origin":
			if !isDowngrade {
				if isSameOrigin {
					httpReq.Header.Set("Referer", fromURL)
				} else {
					httpReq.Header.Set("Referer", originURL)
				}
			}
		case "unsafe-url":
			httpReq.Header.Set("Referer", fromURL)
		default:
			// Empty or unrecognized — browser default: strict-origin-when-cross-origin
			if !isDowngrade {
				if isSameOrigin {
					httpReq.Header.Set("Referer", fromURL)
				} else {
					httpReq.Header.Set("Referer", originURL)
				}
			}
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
