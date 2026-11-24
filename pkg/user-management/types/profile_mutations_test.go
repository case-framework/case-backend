package types

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAddProfileRejectsMainProfileFlag(t *testing.T) {
	main := Profile{ID: primitive.NewObjectID(), MainProfile: true}
	user := User{Profiles: []Profile{main}}
	incoming := Profile{Alias: "Additional", AvatarID: "default", MainProfile: true}

	user.AddProfile(&incoming)

	if incoming.MainProfile || user.Profiles[1].MainProfile {
		t.Fatal("additional profile must not become a main profile")
	}
	if user.Profiles[0] != main {
		t.Fatal("existing main profile changed")
	}
}

func TestUpdateProfilePreservesCreatedAt(t *testing.T) {
	for _, sent := range []int64{0, 1, 9999999999} {
		original := Profile{ID: primitive.NewObjectID(), CreatedAt: 1700000000, MainProfile: true}
		user := User{Profiles: []Profile{original}}
		incoming := Profile{ID: original.ID, Alias: "Updated", AvatarID: "baby", CreatedAt: sent}

		if err := user.UpdateProfile(incoming); err != nil {
			t.Fatal(err)
		}
		got := user.Profiles[0]
		if got.CreatedAt != original.CreatedAt || got.MainProfile != original.MainProfile {
			t.Fatalf("sent createdAt %d: protected fields changed: %+v", sent, got)
		}
		if got.Alias != incoming.Alias || got.AvatarID != incoming.AvatarID {
			t.Fatal("editable fields were not updated")
		}
	}
}

func TestProfileConsentTimes(t *testing.T) {
	for _, tc := range []struct {
		name         string
		stored, sent int64
		renew        bool
	}{
		{"missing", 1700000000, 0, false},
		{"negative", 1700000000, -1, false},
		{"older", 1700000000, 1600000000, false},
		{"equal", 1700000000, 1700000000, false},
		{"same in milliseconds", 1700000000, 1700000000000, false},
		{"stored milliseconds", 1700000000000, 1700000000, false},
		{"older milliseconds", 1700000000, 1600000000000, false},
		{"new seconds", 1700000000, 1800000000, true},
		{"new milliseconds", 1700000000000, 1800000000000, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := primitive.NewObjectID()
			u := User{Profiles: []Profile{{ID: id, ConsentConfirmedAt: tc.stored}}}
			before := time.Now().Unix()
			if err := u.UpdateProfile(Profile{ID: id, ConsentConfirmedAt: tc.sent}); err != nil {
				t.Fatal(err)
			}
			got := u.Profiles[0].ConsentConfirmedAt
			if tc.renew {
				if got < before || got > time.Now().Unix() {
					t.Fatalf("expected server time, got %d", got)
				}
			} else if got != tc.stored {
				t.Fatalf("stored consent changed: %d", got)
			}
		})
	}
}

func TestAddProfileUsesServerTimestamps(t *testing.T) {
	u := User{}
	p := Profile{CreatedAt: 1, ConsentConfirmedAt: 9999999999999, MainProfile: true}
	before := time.Now().Unix()
	u.AddProfile(&p)
	after := time.Now().Unix()
	if p.CreatedAt < before || p.CreatedAt > after || p.ConsentConfirmedAt < before || p.ConsentConfirmedAt > after {
		t.Fatalf("expected server timestamps: %+v", p)
	}
}

func TestUpdateProfileRejectsUnknownProfile(t *testing.T) {
	original := Profile{ID: primitive.NewObjectID(), Alias: "Original"}
	u := User{Profiles: []Profile{original}}
	if err := u.UpdateProfile(Profile{ID: primitive.NewObjectID(), Alias: "Changed"}); err == nil {
		t.Fatal("unknown ID accepted")
	}
	if u.Profiles[0] != original {
		t.Fatal("existing profile changed")
	}
}
