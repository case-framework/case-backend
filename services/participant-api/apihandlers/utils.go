package apihandlers

import (
	"log/slog"
	"math/rand"
	"time"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"

	emailTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

func (h *HttpEndpoints) isInstanceAllowed(instanceID string) bool {
	for _, id := range h.allowedInstanceIDs {
		if id == instanceID {
			return true
		}
	}
	return false
}

func (h *HttpEndpoints) prepAndSendEmailVerification(
	userID string,
	instanceID string,
	email string,
	lang string,
	expiresIn time.Duration,
) {
	tempTokenInfos := userTypes.TempToken{
		UserID:     userID,
		InstanceID: instanceID,
		Purpose:    userTypes.TOKEN_PURPOSE_CONTACT_VERIFICATION,
		Info: map[string]string{
			"type":  userTypes.ACCOUNT_TYPE_EMAIL,
			"email": email,
		},
		Expiration: umUtils.GetExpirationTime(expiresIn),
	}
	tempToken, err := h.globalInfosDBConn.AddTempToken(tempTokenInfos)
	if err != nil {
		slog.Error("failed to create verification token", slog.String("error", err.Error()))
		return
	}

	err = emailsending.SendInstantEmailByTemplate(
		instanceID,
		[]string{email},
		emailTypes.EMAIL_TYPE_REGISTRATION,
		"",
		lang,
		map[string]string{
			"token": tempToken,
		},
		false,
	)
	if err != nil {
		slog.Error("failed to send verification email", slog.String("error", err.Error()))
		return
	}
	slog.Debug("verification email sent", slog.String("email", email))
}

func randomWait(maxTimeSec int) {
	time.Sleep(time.Duration(rand.Intn(maxTimeSec)) * time.Second)
}
