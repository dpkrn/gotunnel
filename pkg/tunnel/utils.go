package tunnel

import "github.com/google/uuid"

func generateConnectionID() string {
	return "conn_" + uuid.New().String()
}

func generateRequestID() string {
	return "req_" + uuid.New().String()
}
