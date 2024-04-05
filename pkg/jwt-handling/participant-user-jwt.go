package jwthandling

import (
	"fmt"
	"time"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"github.com/golang-jwt/jwt/v5"
)

// Information a token enocodes
type ParticipantUserClaims struct {
	InstanceID       string               `json:"instance_id,omitempty"`
	ProfileID        string               `json:"profile_id,omitempty"`
	Payload          map[string]string    `json:"payload,omitempty"`
	AccountConfirmed bool                 `json:"accountConfirmed,omitempty"`
	TempTokenInfos   *userTypes.TempToken `json:"temptoken,omitempty"`
	OtherProfileIDs  []string             `json:"other_profile_ids,omitempty"`
	jwt.RegisteredClaims
}

func GenerateNewParticipantUserToken(
	expiresIn time.Duration,
	id string,
	instanceID string,
	profileID string,
	payload map[string]string,
	accountConfirmed bool,
	tempTokenInfos *userTypes.TempToken,
	otherProfileIDs []string,
	secretKey string,
) (tokenString string, err error) {
	claims := ParticipantUserClaims{
		instanceID,
		profileID,
		payload,
		accountConfirmed,
		tempTokenInfos,
		otherProfileIDs,
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

func ValidateParticipantUserToken(tokenString string, secretKey string) (claims *ParticipantUserClaims, valid bool, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &ManagementUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if token == nil {
		return
	}
	claims, valid = token.Claims.(*ParticipantUserClaims)
	valid = valid && token.Valid
	return
}
