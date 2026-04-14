package tunnel

// DefaultTunnelOptions returns the settings used when [StartTunnel] is called with no options.
func DefaultTunnelOptions() TunnelOptions {
	return TunnelOptions{
		Inspector: true,
		Themes:    string(ThemesDark),
		Logs:      defaultMaxRequestLogs,
	}
}

// applyTunnelOptions merges opts[0] onto [DefaultTunnelOptions]; only non-empty / positive fields override.
func applyTunnelOptions(opts ...TunnelOptions) TunnelOptions {
	o := DefaultTunnelOptions()
	if len(opts) == 0 {
		return o
	}
	u := opts[0]
	if u.Themes != "" {
		o.Themes = u.Themes
	}
	if u.Logs > 0 {
		o.Logs = u.Logs
	}
	if u.InspectorAddr != "" {
		o.InspectorAddr = u.InspectorAddr
	}
	if u.Inspector {
		o.Inspector = true
	}
	return o
}
