package permissionchecker

import (
	"encoding/json"

	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
)

const (
	ManagementUserSubject = "management-user"
	ServiceAccountSubject = "service-account"
)

const (
	RESOURCE_STUDY     = "study"
	RESOURCE_MESSAGING = "messaging"
)

const (
	ACTION_CREATE_STUDY = "create-study"
)

type MuDBConnector interface {
	GetPermissionBySubjectAndResourceForAction(instanceID string, subjectID string, subjectType string, resourceType string, resourceKeys []string, action string) ([]*muDB.Permission, error)
}

func IsAuthorized(db MuDBConnector,
	isAdmin bool,
	instanceID string,
	subjectID string,
	subjectType string,
	resourceType string,
	resourceKeys []string,
	action string,
	infoForLimiter map[string]string,
) bool {
	if isAdmin {
		return true
	}

	permissions, err := getRelevantPermissions(db, instanceID, subjectID, subjectType, resourceType, resourceKeys, action)
	if err != nil {
		return false
	}

	// if there are no permissions, then the user is not authorized
	if len(permissions) == 0 {
		return false
	}

	for _, permission := range permissions {
		// check if the limiter matches the infoForLimiter - if at least one permission matches, then the user is authorized
		if checkLimiter(permission, infoForLimiter) {
			return true
		}
	}

	return false
}

func getRelevantPermissions(db MuDBConnector, instanceID string, subjectID string, subjectType string, resourceType string, resourceKeys []string, action string) ([]*muDB.Permission, error) {
	permissions, err := db.GetPermissionBySubjectAndResourceForAction(instanceID, subjectID, subjectType, resourceType, resourceKeys, action)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func checkLimiter(permission *muDB.Permission, infoForLimiter map[string]string) bool {
	// if the limiter is empty or action does not use a limiter, then it is not limited
	if permission.Limiter == "" || infoForLimiter == nil {
		return true
	}

	// parse the limiter string into a map
	var limiters []map[string]string
	err := parseLimiter(permission.Limiter, &limiters)
	if err != nil {
		return false
	}

	// iterate over the limiters and compare with the infoForLimiter
	for _, limiter := range limiters {
		if compareLimiter(infoForLimiter, limiter) {
			return true
		}
	}

	return false
}

func parseLimiter(limiter string, limiterMap *[]map[string]string) error {
	// parse the string into a map
	err := json.Unmarshal([]byte(limiter), limiterMap)
	if err != nil {
		return err
	}
	return nil
}

func compareLimiter(infoForLimiter map[string]string, limiter map[string]string) bool {
	// iterate over the map and compare the values
	for k, v := range infoForLimiter {
		if limiter[k] != v {
			return false
		}
	}
	return true
}
