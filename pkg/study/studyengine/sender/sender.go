package sender

import (
	"errors"
	"log/slog"
	"maps"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	participantuser "github.com/case-framework/case-backend/pkg/db/participant-user"
	studydb "github.com/case-framework/case-backend/pkg/db/study"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
)

// StudyMessageSender implements studyengine.StudyMessageSender using the platform services.
type StudyMessageSender struct {
	studyDB           *studydb.StudyDBService
	participantUserDB *participantuser.ParticipantUserDBService
	messagingDB       *messagingDB.MessagingDBService
	globalInfosDB     *globalinfosDB.GlobalInfosDBService
}

func NewStudyMessageSender(
	studyDB *studydb.StudyDBService,
	participantUserDB *participantuser.ParticipantUserDBService,
	messagingDB *messagingDB.MessagingDBService,
	globalInfosDB *globalinfosDB.GlobalInfosDBService,
) *StudyMessageSender {
	return &StudyMessageSender{
		studyDB:           studyDB,
		participantUserDB: participantUserDB,
		messagingDB:       messagingDB,
		globalInfosDB:     globalInfosDB,
	}
}

// SendInstantStudyEmail prepares and sends an email immediately using a study template.
func (s *StudyMessageSender) SendInstantStudyEmail(
	instanceID string,
	studyKey string,
	confidentialPID string,
	messageType string,
	extraPayload map[string]string,
	opts studyengine.SendOptions,
) error {
	if s.studyDB == nil || s.participantUserDB == nil || s.messagingDB == nil {
		return errors.New("sender not initialized correctly")
	}

	// Map confidential participant id to profile id
	profileID, err := s.studyDB.GetProfileIDFromConfidentialID(instanceID, confidentialPID, studyKey)
	if err != nil || profileID == "" {
		slog.Error("profileID lookup failed", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("confidentialPID", confidentialPID))
		return errors.New("profileID lookup failed")
	}

	// Load user
	user, err := s.participantUserDB.GetUserByProfileID(instanceID, profileID)
	if err != nil {
		slog.Error("user lookup failed", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("profileID", profileID), slog.String("error", err.Error()))
		return err
	}

	// Determine recipient and profile info
	email, err := user.GetEmail()
	if err != nil {
		slog.Error("email lookup failed", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("profileID", profileID), slog.String("error", err.Error()))
		return err
	}
	to := []string{email.Email}

	currentProfile := user.Profiles[0]
	for _, p := range user.Profiles {
		if p.ID.Hex() == profileID {
			currentProfile = p
			break
		}
	}

	// Build payload
	payload := map[string]string{
		"studyKey":     studyKey,
		"profileAlias": currentProfile.Alias,
		"profileId":    currentProfile.ID.Hex(),
	}
	// Merge extra payload (action-provided)
	maps.Copy(payload, extraPayload)

	// Determine language
	lang := user.Account.PreferredLanguage
	if opts.LanguageOverride != "" {
		lang = opts.LanguageOverride
	}

	// Send immediately using the templating system; default to high priority
	err = emailsending.SendInstantEmailByTemplate(
		instanceID,
		to,
		messageType,
		studyKey,
		lang,
		payload,
		false, // useLowPrio
		opts.ExpiresAt,
	)
	if err != nil {
		// The email-sending module already stores to outgoing on error
		return err
	}
	return nil
}
