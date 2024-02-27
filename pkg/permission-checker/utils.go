package permissionchecker

import (
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
	if permission.Limiter == nil || infoForLimiter == nil {
		return true
	}

	// iterate over the limiters and compare with the infoForLimiter
	for _, limiter := range permission.Limiter {
		if compareLimiter(infoForLimiter, limiter) {
			return true
		}
	}

	return false
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
