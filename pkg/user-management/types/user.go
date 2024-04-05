package participantuser

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ACCOUNT_TYPE_EMAIL = "email"

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	Account            Account            `bson:"account" json:"account"`
	Timestamps         Timestamps         `bson:"timestamps" json:"timestamps"`
	Profiles           []Profile          `bson:"profiles" json:"profiles"`
	ContactPreferences ContactPreferences `bson:"contactPreferences" json:"contactPreferences"`
	ContactInfos       []ContactInfo
}

// Add a new email address
func (u *User) AddNewEmail(addr string, confirmed bool) {
	contactInfo := ContactInfo{
		ID:          primitive.NewObjectID(),
		Type:        "email",
		ConfirmedAt: 0,
		Email:       addr,
	}
	if confirmed {
		contactInfo.ConfirmedAt = time.Now().Unix()
	}
	u.ContactInfos = append(u.ContactInfos, contactInfo)
}

func (u *User) ConfirmContactInfo(t string, addr string) error {
	for i, ci := range u.ContactInfos {
		if t == "email" && ci.Email == addr {
			u.ContactInfos[i].ConfirmedAt = time.Now().Unix()
			return nil
		} else if t == "phone" && ci.Phone == addr {
			u.ContactInfos[i].ConfirmedAt = time.Now().Unix()
			return nil
		}
	}
	return errors.New("contact not found")
}

func (u *User) SetContactInfoVerificationSent(t string, addr string) {
	for i, ci := range u.ContactInfos {
		if t == "email" && ci.Email == addr {
			u.ContactInfos[i].ConfirmationLinkSentAt = time.Now().Unix()
			return
		} else if t == "phone" && ci.Phone == addr {
			u.ContactInfos[i].ConfirmationLinkSentAt = time.Now().Unix()
			return
		}
	}
}

func (u User) FindContactInfoByTypeAndAddr(t string, addr string) (ContactInfo, bool) {
	for _, ci := range u.ContactInfos {
		if t == "email" && ci.Email == addr {
			return ci, true
		} else if t == "phone" && ci.Phone == addr {
			return ci, true
		}
	}
	return ContactInfo{}, false
}

func (u User) FindContactInfoById(id string) (ContactInfo, bool) {
	for _, ci := range u.ContactInfos {
		if ci.ID.Hex() == id {
			return ci, true
		}
	}
	return ContactInfo{}, false
}

// RemoveContactInfo from the user and also all references from the contact preferences
func (u *User) RemoveContactInfo(id string) error {
	for i, ci := range u.ContactInfos {
		if ci.ID.Hex() == id {
			if u.Account.Type == "email" && ci.Email == u.Account.AccountID {
				return errors.New("cannot remove main address")
			}

			u.ContactInfos = append(u.ContactInfos[:i], u.ContactInfos[i+1:]...)
			return nil
		}
	}
	u.RemoveContactInfoFromContactPreferences(id)
	return errors.New("contact not found")
}

// RemoveContactInfoFromContactPreferences should delete all references to a contact info object
func (u *User) RemoveContactInfoFromContactPreferences(id string) {
	// remove address from contact preferences
	for i, addrRef := range u.ContactPreferences.SendNewsletterTo {
		if addrRef == id {
			u.ContactPreferences.SendNewsletterTo = append(u.ContactPreferences.SendNewsletterTo[:i], u.ContactPreferences.SendNewsletterTo[i+1:]...)
			return
		}
	}
}

// ReplaceContactInfoInContactPreferences to use if a new contact reference should replace to old one
func (u *User) ReplaceContactInfoInContactPreferences(oldId string, newId string) {
	// replace address from contact preferences
	for i, addrRef := range u.ContactPreferences.SendNewsletterTo {
		if addrRef == oldId {
			u.ContactPreferences.SendNewsletterTo[i] = newId
		}
	}
}

// AddProfile generates unique ID and adds profile to the user's array
func (u *User) AddProfile(p Profile) {
	p.ID = primitive.NewObjectID()
	p.CreatedAt = time.Now().Unix()
	u.Profiles = append(u.Profiles, p)
}

// UpdateProfile finds and replaces profile in the user's array
func (u *User) UpdateProfile(p Profile) error {
	for i, cP := range u.Profiles {
		if cP.ID == p.ID {
			p.MainProfile = cP.MainProfile
			u.Profiles[i] = p
			return nil
		}
	}
	return errors.New("profile with given ID not found")
}

// FindProfile finds a profile in the user's array
func (u User) FindProfile(id string) (Profile, error) {
	for _, cP := range u.Profiles {
		if cP.ID.Hex() == id {
			return cP, nil
		}
	}
	return Profile{}, errors.New("profile with given ID not found")
}

// RemoveProfile finds and removes profile from the user's array
func (u *User) RemoveProfile(id string) error {
	for i, cP := range u.Profiles {
		if cP.ID.Hex() == id {
			if cP.MainProfile {
				return errors.New("cannot remove main profile")
			}
			u.Profiles = append(u.Profiles[:i], u.Profiles[i+1:]...)
			return nil
		}
	}
	return errors.New("profile with given ID not found")
}

type Timestamps struct {
	LastTokenRefresh        int64 `bson:"lastTokenRefresh" json:"lastTokenRefresh"`
	LastLogin               int64 `bson:"lastLogin" json:"lastLogin"`
	CreatedAt               int64 `bson:"createdAt" json:"createdAt"`
	UpdatedAt               int64 `bson:"updatedAt" json:"updatedAt"`
	LastPasswordChange      int64 `bson:"lastPasswordChange" json:"lastPasswordChange"`
	ReminderToConfirmSentAt int64 `bson:"reminderToConfirmSentAt" json:"reminderToConfirmSentAt"`
	MarkedForDeletion       int64 `bson:"markedForDeletion" json:"markedForDeletion"`
}
