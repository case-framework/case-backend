package apihandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	"github.com/gin-gonic/gin"
)

func TestAddStudyVariableRequestAllowsIdentifierKeys(t *testing.T) {
	for _, key := range []string{"isSummerPause", "toggleForSwabbing", "variable_1", "variable-1", "123", "_-"} {
		t.Run(key, func(t *testing.T) {
			req := AddStudyVariableRequest{VariableDef: studyTypes.StudyVariables{
				Key:         key,
				Type:        studyTypes.STUDY_VARIABLES_TYPE_STRING,
				Label:       "<img src=x onerror=alert(1)></img>",
				Description: "Description with spaces / punctuation",
				Value:       "<img src=x onerror=alert(1)></img>",
			}}
			if err := req.validate(); err != nil {
				t.Fatalf("valid key %q rejected: %v", key, err)
			}
		})
	}
}

func TestAddStudyVariableRejectsInvalidKeysBeforeDatabaseAccess(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{name: "reported HTML payload", key: "<img src=x onerror=alert(1)></img>"},
		{name: "empty", key: ""},
		{name: "space", key: "some variable"},
		{name: "leading space", key: " variable"},
		{name: "trailing newline", key: "variable\n"},
		{name: "slash", key: "some/variable"},
		{name: "backslash", key: `some\variable`},
		{name: "query", key: "variable?value=1"},
		{name: "fragment", key: "variable#value"},
		{name: "encoded slash", key: "some%2Fvariable"},
		{name: "dot", key: "."},
		{name: "parent segment", key: ".."},
		{name: "unicode letter", key: "variäble"},
	}

	// No database is configured: invalid requests must return before accessing it.
	handler := &HttpEndpoints{}
	router := gin.New()
	router.POST("/studies/:studyKey/variables", func(c *gin.Context) {
		c.Set("validatedToken", &jwthandling.ManagementUserClaims{InstanceID: "test"})
		handler.addStudyVariable(c)
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(AddStudyVariableRequest{VariableDef: studyTypes.StudyVariables{
				Key:  tc.key,
				Type: studyTypes.STUDY_VARIABLES_TYPE_STRING,
			}})
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/studies/test/variables", bytes.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected HTTP 400, got %d: %s", response.Code, response.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error != "variable key must contain only ASCII letters, digits, underscores, or hyphens and must not be empty" {
				t.Fatalf("unexpected validation error: %q", body.Error)
			}
		})
	}
}
