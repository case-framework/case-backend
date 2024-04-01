package utils

import (
	"testing"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetMainAndOtherProfiles(t *testing.T) {
	t.Run("with a single profile with main flag", func(t *testing.T) {
		user := userTypes.User{
			Profiles: []userTypes.Profile{
				{ID: primitive.NewObjectID(), MainProfile: true},
			},
		}
		main, others := GetMainAndOtherProfiles(user)
		if main != user.Profiles[0].ID.Hex() {
			t.Errorf("unexpected main id %s", main)
		}
		if len(others) != 0 {
			t.Errorf("unexpected number of other profiles %d", len(others))
		}
	})

	t.Run("with a single profile without main flag", func(t *testing.T) {
		user := userTypes.User{
			Profiles: []userTypes.Profile{
				{ID: primitive.NewObjectID()},
			},
		}
		main, others := GetMainAndOtherProfiles(user)
		if main != user.Profiles[0].ID.Hex() {
			t.Errorf("unexpected main id %s", main)
		}
		if len(others) != 0 {
			t.Errorf("unexpected number of other profiles %d", len(others))
		}
	})

	t.Run("with mulitple profiles without main flag", func(t *testing.T) {
		user := userTypes.User{
			Profiles: []userTypes.Profile{
				{ID: primitive.NewObjectID(), MainProfile: false},
				{ID: primitive.NewObjectID(), MainProfile: false},
				{ID: primitive.NewObjectID(), MainProfile: false},
				{ID: primitive.NewObjectID(), MainProfile: false},
			},
		}
		main, others := GetMainAndOtherProfiles(user)
		if main != user.Profiles[0].ID.Hex() {
			t.Errorf("unexpected main id %s", main)
		}
		if len(others) != 3 || others[0] == main {
			t.Errorf("unexpected number of other profiles %d or wrong ids", len(others))
		}
	})

	t.Run("with mulitple profiles one main flag", func(t *testing.T) {
		user := userTypes.User{
			Profiles: []userTypes.Profile{
				{ID: primitive.NewObjectID(), MainProfile: false},
				{ID: primitive.NewObjectID(), MainProfile: true},
				{ID: primitive.NewObjectID(), MainProfile: false},
			},
		}
		main, others := GetMainAndOtherProfiles(user)
		if main != user.Profiles[1].ID.Hex() {
			t.Errorf("unexpected main id %s", main)
		}
		if len(others) != 2 {
			t.Errorf("unexpected number of other profiles %d", len(others))
		}
	})

	t.Run("with mulitple profiles multiply main flag", func(t *testing.T) {
		user := userTypes.User{
			Profiles: []userTypes.Profile{
				{ID: primitive.NewObjectID(), MainProfile: false},
				{ID: primitive.NewObjectID(), MainProfile: true},
				{ID: primitive.NewObjectID(), MainProfile: true},
			},
		}
		main, others := GetMainAndOtherProfiles(user)
		if main != user.Profiles[2].ID.Hex() {
			t.Errorf("unexpected main id %s", main)
		}
		if len(others) != 1 {
			t.Errorf("unexpected number of other profiles %d", len(others))
		}
	})

}
