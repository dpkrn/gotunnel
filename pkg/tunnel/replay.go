package tunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type replayPayload struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// replayResult is the JSON returned after replaying a request against the local app.
type replayResult struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
	Error   string              `json:"error,omitempty"`
}

// Hop-by-hop and connection-specific headers we must not copy when replaying.
var replayHeaderBlocklist = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {},
	"Trailers":            {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
	"Host":                {},
}

func handleReplay(localPort string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(replayResult{Error: "use POST"})
			return
		}

		var payload replayPayload
		dec := json.NewDecoder(io.LimitReader(r.Body, 10<<20))
		if err := dec.Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(replayResult{Error: "invalid JSON: " + err.Error()})
			return
		}

		method := strings.ToUpper(strings.TrimSpace(payload.Method))
		if method == "" {
			method = http.MethodGet
		}

		path := strings.TrimSpace(payload.Path)
		if path == "" {
			path = "/"
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		target := fmt.Sprintf("http://127.0.0.1:%s%s", localPort, path)
		req, err := http.NewRequest(method, target, strings.NewReader(payload.Body))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(replayResult{Error: err.Error()})
			return
		}

		if payload.Headers != nil {
			for k, vals := range payload.Headers {
				can := http.CanonicalHeaderKey(k)
				if _, skip := replayHeaderBlocklist[can]; skip {
					continue
				}
				for _, v := range vals {
					req.Header.Add(can, v)
				}
			}
		}

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(replayResult{Error: err.Error()})
			return
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(replayResult{Error: "read response: " + err.Error()})
			return
		}

		out := replayResult{
			Status:  resp.StatusCode,
			Headers: map[string][]string(resp.Header),
			Body:    string(b),
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(out)
	}
}
