package client

import "fmt"

func withCredential(headers map[string]string, accessToken string) map[string]string {
	headers["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
	return headers
}
