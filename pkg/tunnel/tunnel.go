package tunnel

import (
	"fmt"

	"github.com/DpkRn/gotunnel/internal/tunnel"
)

func StartTunnel(port string) (url string, stop func(), err error) {

	tunnel, err := tunnel.NewTunnel(port)
	if err != nil {
		return "", noop, fmt.Errorf("could not create tunnel %w", err)
	}

	go func() {
		err = tunnel.Start()
		if err != nil {
			fmt.Println("could not start tunnel", err)
		}
	}()

	publicURL := tunnel.GetPublicUrl()
	stop = func() {
		tunnel.Stop()
	}

	return publicURL, stop, nil
}
func noop() {}
