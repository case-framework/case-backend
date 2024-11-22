package jwthandling

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Information a token enocodes
type ManagementUserClaims struct {
	InstanceID    string            `json:"instance_id,omitempty"`
	IsAdmin       bool              `json:"is_admin,omitempty"`
	IsServiceUser bool              `json:"is_service_user"`
	Payload       map[string]string `json:"payload,omitempty"`
	jwt.RegisteredClaims
}

func GenerateNewManagementUserToken(expiresIn time.Duration, id string, instanceID string, isAdmin bool, payload map[string]string, secretKey string) (tokenString string, err error) {
	claims := ManagementUserClaims{
		instanceID,
		isAdmin,
		false,
		payload,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(secretKey))
	return
}

func ValidateManagementUserToken(tokenString string, secretKey string) (claims *ManagementUserClaims, valid bool, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &ManagementUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if token == nil {
		return
	}
	claims, valid = token.Claims.(*ManagementUserClaims)
	valid = valid && token.Valid
	return
}
