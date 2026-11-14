package labtest

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Notice is one request to the CDR arrangement revocation endpoint of the JWKS server.
type Notice struct {
	ContentType   string
	Authorization string
	Form          url.Values
}

var notices = struct {
	mu        sync.Mutex
	received  map[string][]Notice
	responses map[string]int
}{received: map[string][]Notice{}, responses: map[string]int{}}

// RespondToNotices sets the status of the response to each notice for the arrangement. The default is 204.
func RespondToNotices(arrangementID string, status int) {
	notices.mu.Lock()
	defer notices.mu.Unlock()

	notices.responses[arrangementID] = status
}

// NoticesFor gives the notices that the JWKS server got for the arrangement, in the sequence of their arrival.
func NoticesFor(arrangementID string) []Notice {
	notices.mu.Lock()
	defer notices.mu.Unlock()

	return append([]Notice(nil), notices.received[arrangementID]...)
}

// receiveNotice finds the arrangement in the cdr_arrangement_jwt with no check of its signature: the test checks it.
func receiveNotice(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	arrangementID := unverifiedArrangementID(r.PostForm.Get("cdr_arrangement_jwt"))

	notices.mu.Lock()
	notices.received[arrangementID] = append(notices.received[arrangementID], Notice{
		ContentType:   r.Header.Get("Content-Type"),
		Authorization: r.Header.Get("Authorization"),
		Form:          r.PostForm,
	})

	status, ok := notices.responses[arrangementID]
	notices.mu.Unlock()

	if !ok {
		status = http.StatusNoContent
	}

	w.WriteHeader(status)
}

func unverifiedArrangementID(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}

	var claims struct {
		ArrangementID string `json:"cdr_arrangement_id"`
	}

	_ = json.Unmarshal(payload, &claims)

	return claims.ArrangementID
}
