package apihandlers

import (
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

func randomWait(maxTimeSec int) {
	time.Sleep(time.Duration(rand.Intn(maxTimeSec)) * time.Second)
}
