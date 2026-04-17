package tunnel

// Options configures the tunnel client.
type Options struct {
	// Inspector enables sending captured traffic to the inspector ingest WebSocket (default true when nil).
	Inspector *bool

	// InspectorPort is the inspector listen port or host:port (e.g. "4040", ":9090") for both the
	// embedded server (Go) and the ws://…/ingest URL.
	InspectorPort string

	// EmbeddedInspector starts the inspector HTTP server in-process before dialing (Go only, default true when nil).
	// Set false if you run cmd/inspector or your own server on this port. Nodetunnel cannot embed Go;
	// use js/spawnInspector.mjs or spawn the inspector binary yourself.
	EmbeddedInspector *bool
}

// Option mutates Options when passed to [StartTunnel].
type Option func(*Options)

func WithInspector(enabled bool) Option {
	return func(o *Options) {
		o.Inspector = &enabled
	}
}

func WithInspectorPort(port string) Option {
	return func(o *Options) {
		o.InspectorPort = port
	}
}

func WithEmbeddedInspector(start bool) Option {
	return func(o *Options) {
		o.EmbeddedInspector = &start
	}
}

func inspectorEnabled(o *Options) bool {
	if o.Inspector != nil {
		return *o.Inspector
	}
	return true
}

func embeddedInspectorEnabled(o *Options) bool {
	if !inspectorEnabled(o) {
		return false
	}
	if o.EmbeddedInspector != nil {
		return *o.EmbeddedInspector
	}
	return true
}
