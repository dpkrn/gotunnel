module github.com/dpkrn/gotunnel/mytunnel

go 1.25.0

require (
	github.com/dpkrn/gotunnel v0.3.6
	github.com/google/uuid v1.6.0
)

// Local dev: go.work also replaces this. Remove after tagging v0.4.0+ with module github.com/dpkrn/gotunnel.

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
)
