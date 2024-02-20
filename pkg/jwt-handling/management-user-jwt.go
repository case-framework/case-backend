package jwthandling

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

// Information a token enocodes
type ManagementUserClaims struct {
	ID         string            `json:"id,omitempty"`
	InstanceID string            `json:"instance_id,omitempty"`
	IsAdmin    bool              `json:"is_admin,omitempty"`
	Payload    map[string]string `json:"payload,omitempty"`
	jwt.RegisteredClaims
}

func ValidateManagementUserToken(tokenString string, secretKey string) (claims *ManagementUserClaims, valid bool, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &ManagementUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if token == nil {
		return
	}
	claims, valid = token.Claims.(*ManagementUserClaims)
	valid = valid && token.Valid
	return
}
