// jwt.go
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

type Payload struct {
	APIKey string `json:"api_key"`
}

func DecodeAndVerifyPayload(token, secret string) (Payload, error) {
	var payload Payload

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return payload, errors.New("invalid token format")
	}

	decodedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return payload, err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(decodedSignature, expectedSignature) {
		return payload, errors.New("invalid signature")
	}

	decodedPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return payload, err
	}

	err = json.Unmarshal(decodedPayload, &payload)
	if err != nil {
		return payload, err
	}

	return payload, nil
}
