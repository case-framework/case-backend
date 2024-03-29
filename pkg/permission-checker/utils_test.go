package permissionchecker

import (
	"testing"

	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
)

type mockMuDBConnector struct {
	permissions []*muDB.Permission
}

func (m *mockMuDBConnector) GetPermissionBySubjectAndResourceForAction(instanceID string, subjectID string, subjectType string, resourceType string, resourceKeys []string, action string) ([]*muDB.Permission, error) {
	// return permissions after filtering
	filteredPermissions := make([]*muDB.Permission, 0)
	actions := []string{action}
	if action != "*" {
		actions = append(actions, "*")
	}

	for _, p := range m.permissions {
		if p.SubjectID == subjectID && p.SubjectType == subjectType && p.ResourceType == resourceType {
			for _, a := range actions {
				if p.Action == a {
					for _, key := range resourceKeys {
						if p.ResourceKey == key {
							filteredPermissions = append(filteredPermissions, p)
							break
						}
					}
				}
			}
		}
	}
	return filteredPermissions, nil
}

func TestIsAuthorized(t *testing.T) {
	t.Parallel()

	mockMuDBConnector := &mockMuDBConnector{
		permissions: []*muDB.Permission{
			{
				SubjectID:    "sub1",
				SubjectType:  "2",
				ResourceType: "3",
				ResourceKey:  "4",
				Action:       "5",
				Limiter:      nil,
			},
			{
				SubjectID:    "sub3",
				SubjectType:  "2",
				ResourceType: "3",
				ResourceKey:  "4",
				Action:       "5",
				Limiter: []map[string]string{
					{"a": "1", "b": "2"},
				},
			},
			{
				SubjectID:    "sub4",
				SubjectType:  "2",
				ResourceType: "3",
				ResourceKey:  "4",
				Action:       "5",
				Limiter: []map[string]string{
					{"a": "1", "b": "2"},
				},
			},
			{
				SubjectID:    "sub4",
				SubjectType:  "2",
				ResourceType: "3",
				ResourceKey:  "*",
				Action:       "5",
				Limiter:      nil,
			},
		},
	}

	tests := []struct {
		isAdmin        bool
		subjectID      string
		subjectType    string
		resourceType   string
		resourceKeys   []string
		action         string
		infoForLimiter map[string]string
		expected       bool
	}{
		// isAdmin = true:
		{
			isAdmin:        true,
			subjectID:      "sub1",
			subjectType:    "2",
			resourceType:   "3",
			resourceKeys:   []string{"4"},
			action:         "5",
			infoForLimiter: nil,
			expected:       true,
		},
		// isAdmin = false, no permissions:
		{
			isAdmin:        false,
			subjectID:      "sub1",
			subjectType:    "2",
			resourceType:   "not found",
			resourceKeys:   []string{"4"},
			action:         "5",
			infoForLimiter: nil,
			expected:       false,
		},
		// isAdmin = false, has permissions with no limiter:
		{
			isAdmin:        false,
			subjectID:      "sub1",
			subjectType:    "2",
			resourceType:   "3",
			resourceKeys:   []string{"4"},
			action:         "5",
			infoForLimiter: map[string]string{"key": "ignored"},
			expected:       true,
		},
		// isAdmin = false, has permissions with correct limiter format but not matching limiter info
		{
			isAdmin:        false,
			subjectID:      "sub3",
			subjectType:    "2",
			resourceType:   "3",
			resourceKeys:   []string{"4"},
			action:         "5",
			infoForLimiter: map[string]string{"a": "1", "b": "3"},
			expected:       false,
		},
		// isAdmin = false, has permissions with correct limiter format and matching limiter info
		{
			isAdmin:        false,
			subjectID:      "sub4",
			subjectType:    "2",
			resourceType:   "3",
			resourceKeys:   []string{"4"},
			action:         "5",
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			expected:       true,
		},
		// isAdmin = false, has permissions with one limited other not limited
		{
			isAdmin:        false,
			subjectID:      "sub4",
			subjectType:    "2",
			resourceType:   "3",
			resourceKeys:   []string{"4", "*"},
			action:         "5",
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			expected:       true,
		},
	}

	for index, test := range tests {
		result := IsAuthorized(mockMuDBConnector, test.isAdmin, "instanceID", test.subjectID, test.subjectType, test.resourceType, test.resourceKeys, test.action, test.infoForLimiter)
		if result != test.expected {
			t.Errorf("test %d: expected %t for input %v, %v, %v, %v, %v, %v, %v but got %t", index, test.expected, test.isAdmin, test.subjectID, test.subjectType, test.resourceType, test.resourceKeys, test.action, test.infoForLimiter, result)
		}
	}
}

func TestCheckLimiter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		infoForLimiter map[string]string
		permission     *muDB.Permission
		expected       bool
	}{
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			permission: &muDB.Permission{
				Limiter: []map[string]string{
					{"a": "1", "b": "2"},
				},
			},
			expected: true,
		},
		{
			infoForLimiter: nil,
			permission: &muDB.Permission{
				Limiter: []map[string]string{
					{"a": "1", "b": "2"},
				},
			},
			expected: false,
		},
		{
			infoForLimiter: nil,
			permission: &muDB.Permission{
				Limiter: nil,
			},
			expected: true,
		},
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			permission: &muDB.Permission{
				Limiter: nil,
			},
			expected: true,
		},
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			permission: &muDB.Permission{
				Limiter: []map[string]string{
					{"a": "1", "b": "3"},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		result := checkLimiter(test.permission, test.infoForLimiter)
		if result != test.expected {
			t.Errorf("expected %t for input %v, %v but got %t", test.expected, test.infoForLimiter, test.permission, result)
		}
	}
}

func TestCompareLimiter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		infoForLimiter map[string]string
		limiter        map[string]string
		expected       bool
	}{
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			limiter:        map[string]string{"a": "1", "b": "2"},
			expected:       true,
		},
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			limiter:        map[string]string{"a": "1", "b": "3"},
			expected:       false,
		},
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			limiter:        map[string]string{"a": "1"},
			expected:       false,
		},
		{
			infoForLimiter: map[string]string{"a": "1", "b": "2"},
			limiter:        map[string]string{"a": "1", "b": "2", "c": "3"},
			expected:       true,
		},
	}

	for _, test := range tests {
		result := compareLimiter(test.infoForLimiter, test.limiter)
		if result != test.expected {
			t.Errorf("expected %t for input %v, %v but got %t", test.expected, test.infoForLimiter, test.limiter, result)
		}
	}
}
