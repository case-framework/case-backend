package utils

import (
	"time"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func InitNewEmailUser(
	email string,
	password string,
	locale string,
) userTypes.User {
	newUser := userTypes.User{
		Account: userTypes.Account{
			Type:               userTypes.ACCOUNT_TYPE_EMAIL,
			AccountID:          email,
			Password:           password,
			AccountConfirmedAt: 0,
			PreferredLanguage:  locale,
		},
		Profiles: []userTypes.Profile{
			{
				ID:                 primitive.NewObjectID(),
				Alias:              BlurEmailAddress(email),
				MainProfile:        true,
				AvatarID:           "default",
				ConsentConfirmedAt: time.Now().Unix(),
			},
		},
		Timestamps: userTypes.Timestamps{
			CreatedAt: time.Now().Unix(),
		},
	}
	newUser.AddNewEmail(email, false)

	newUser.ContactPreferences = userTypes.ContactPreferences{
		SubscribedToWeekly:            true,
		ReceiveWeeklyMessageDayOfWeek: int32(CurrentWeekdayStrategy.Weekday()),
	}

	return newUser
}
