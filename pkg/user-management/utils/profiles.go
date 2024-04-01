package utils

import (
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

// GetMainAndOtherProfiles extract main profile ID and secondary profiles from a user account info
func GetMainAndOtherProfiles(user userTypes.User) (mainProfileID string, otherProfileIDs []string) {
	mainProfileID = ""
	otherProfileIDs = []string{}
	for _, p := range user.Profiles {
		if !p.MainProfile {
			otherProfileIDs = append(otherProfileIDs, p.ID.Hex())
		} else {
			mainProfileID = p.ID.Hex()
		}
	}
	if mainProfileID == "" {
		mainProfileID = otherProfileIDs[0]
		otherProfileIDs = otherProfileIDs[1:]
	}
	return mainProfileID, otherProfileIDs
}
