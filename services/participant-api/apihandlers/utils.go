package apihandlers

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
)

func (h *HttpEndpoints) isInstanceAllowed(instanceID string) bool {
	for _, id := range h.allowedInstanceIDs {
		if id == instanceID {
			return true
		}
	}
	return false
}

func (h *HttpEndpoints) prepTokenAndSendEmail(
	userID string,
	instanceID string,
	email string,
	lang string,
	tokenPurpose string,
	expiresIn time.Duration,
	emailTemplate string,
	payload map[string]string,
) {
	tempTokenInfos := userTypes.TempToken{
		UserID:     userID,
		InstanceID: instanceID,
		Purpose:    tokenPurpose,
		Info: map[string]string{
			"type":  userTypes.ACCOUNT_TYPE_EMAIL,
			"email": email,
		},
		Expiration: umUtils.GetExpirationTime(expiresIn),
	}
	tempToken, err := h.globalInfosDBConn.AddTempToken(tempTokenInfos)
	if err != nil {
		slog.Error("failed to create token", slog.String("error", err.Error()))
		return
	}

	if payload == nil {
		payload = make(map[string]string)
	}
	payload["token"] = tempToken

	err = emailsending.SendInstantEmailByTemplate(
		instanceID,
		[]string{email},
		emailTemplate,
		"",
		lang,
		payload,
		false,
	)
	if err != nil {
		slog.Error("failed to send email", slog.String("error", err.Error()))
		return
	}
	slog.Debug("email sent", slog.String("email", email))

}

func (h *HttpEndpoints) prepAndSendEmailVerification(
	userID string,
	instanceID string,
	email string,
	lang string,
	expiresIn time.Duration,
	emailTemplate string,
) {
	h.prepTokenAndSendEmail(
		userID,
		instanceID,
		email,
		lang,
		userTypes.TOKEN_PURPOSE_CONTACT_VERIFICATION,
		expiresIn,
		emailTemplate,
		nil,
	)
}

func (h *HttpEndpoints) sendSimpleEmail(
	instanceID string, to []string, messageType string, studyKey string, lang string, payload map[string]string, useLowPrio bool,
) {
	err := emailsending.SendInstantEmailByTemplate(
		instanceID,
		to,
		messageType,
		studyKey,
		lang,
		payload,
		useLowPrio,
	)
	if err != nil {
		slog.Error("failed to send email", slog.String("error", err.Error()))
		return
	}
}

func randomWait(minTimeSec int, maxTimeSec int) {
	time.Sleep(time.Duration(rand.Intn(maxTimeSec-minTimeSec)+minTimeSec) * time.Second)
}

func (h *HttpEndpoints) validateTempToken(token string, purposes []string) (tt userTypes.TempToken, err error) {
	tokenInfos, err := h.globalInfosDBConn.GetTempToken(token)
	if err != nil {
		return
	}
	if tokenInfos.Expiration.Before(time.Now()) {
		err = errors.New("token expired")
		return
	}
	for _, purpose := range purposes {
		if tokenInfos.Purpose == purpose {
			return tokenInfos, nil
		}
	}
	err = fmt.Errorf("wrong token purpose: %s", tokenInfos.Purpose)
	return
}

func (h *HttpEndpoints) checkProfileBelongsToUser(instanceID, userID, profileID string) bool {
	user, err := h.userDBConn.GetUser(instanceID, userID)
	if err != nil {
		slog.Warn("user not found", slog.String("instanceID", instanceID), slog.String("userID", userID), slog.String("error", err.Error()))
		return false
	}

	for _, profile := range user.Profiles {
		if profile.ID.Hex() == profileID {
			return true
		}
	}
	return false
}

func (h *HttpEndpoints) checkAllProfilesBelongsToUser(instanceID, userID string, profileIDs []string) bool {
	user, err := h.userDBConn.GetUser(instanceID, userID)
	if err != nil {
		slog.Warn("user not found", slog.String("instanceID", instanceID), slog.String("userID", userID), slog.String("error", err.Error()))
		return false
	}

	for _, profileID := range profileIDs {
		found := false
		for _, profile := range user.Profiles {
			if profile.ID.Hex() == profileID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
