package cds

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// acceptsJSON is true for an absent or empty Accept header, because the CDS then gives application/json.
func acceptsJSON(r *http.Request) bool {
	values := r.Header.Values("Accept")
	if strings.TrimSpace(strings.Join(values, "")) == "" {
		return true
	}

	for _, value := range values {
		for mediaRange := range strings.SplitSeq(value, ",") {
			if permitsJSON(mediaRange) {
				return true
			}
		}
	}

	return false
}

func permitsJSON(mediaRange string) bool {
	mediaType, params, err := mime.ParseMediaType(mediaRange)
	if err != nil {
		return false
	}

	switch mediaType {
	case "*/*", "application/*", "application/json":
	default:
		return false
	}

	if charset, ok := params["charset"]; ok && !strings.EqualFold(charset, "utf-8") {
		return false
	}

	if q, ok := params["q"]; ok {
		weight, err := strconv.ParseFloat(q, 64)
		if err != nil || weight <= 0 {
			return false
		}
	}

	return true
}
