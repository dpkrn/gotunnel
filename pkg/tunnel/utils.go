package tunnel

import "github.com/google/uuid"

func GenerateConnectionID() string {
	return "conn_" + uuid.New().String()
}

func GenerateRequestID() string {
	return "req_" + uuid.New().String()
}
