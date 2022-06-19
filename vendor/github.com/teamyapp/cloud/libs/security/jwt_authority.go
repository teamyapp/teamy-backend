package security

import (
	"encoding/json"
	"errors"

	"github.com/golang-jwt/jwt"
)

type JWTAuthority struct {
	signingKey []byte
}

func (j JWTAuthority) GenerateToken(payload interface{}) (string, error) {
	payloadMap := make(map[string]interface{})
	jsonBuf, _ := json.Marshal(payload)
	err := json.Unmarshal(jsonBuf, &payloadMap)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payloadMap))
	return token.SignedString(j.signingKey)
}

func (j JWTAuthority) DecodeToken(jwtToken string, output interface{}) error {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("token is invalid")
	}

	return parseJWTClaims(token.Claims, output)
}

func (j JWTAuthority) DecodeUnverifiedToken(jwtToken string, output interface{}) error {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &claims)
	if err != nil {
		return err
	}

	return parseJWTClaims(claims, output)
}

func parseJWTClaims(claims jwt.Claims, output interface{}) error {
	buf, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, output)
}

func NewJWTAuthority(signingKey string) JWTAuthority {
	return JWTAuthority{
		signingKey: []byte(signingKey),
	}
}
