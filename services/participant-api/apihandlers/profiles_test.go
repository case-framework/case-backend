package apihandlers

import (
	"encoding/json"
	"fmt"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"github.com/gin-gonic/gin"
)

func TestCreateProfileRejectsInvalidAlias(t *testing.T) {
	// No database is configured: invalid aliases must be rejected before DB access.
	h := &HttpEndpoints{}
	for _, endpoint := range []struct {
		method  string
		handler gin.HandlerFunc
	}{
		{http.MethodPost, h.addNewProfileHandl},
	} {
		for _, tc := range []struct{ name, payload, message string }{
			{"long ASCII", `{"alias":"` + strings.Repeat("a", umUtils.ProfileAliasMaxUTF16Length+1) + `"}`, fmt.Sprintf("profile alias must not exceed %d UTF-16 code units", umUtils.ProfileAliasMaxUTF16Length)},
			{"long emoji", `{"alias":"` + strings.Repeat("😀", umUtils.ProfileAliasMaxUTF16Length/2+1) + `"}`, fmt.Sprintf("profile alias must not exceed %d UTF-16 code units", umUtils.ProfileAliasMaxUTF16Length)},
			{"empty", `{"alias":""}`, "profile alias must not be blank"},
			{"spaces", `{"alias":"   "}`, "profile alias must not be blank"},
			{"Unicode whitespace", `{"alias":"\u00a0\t\n"}`, "profile alias must not be blank"},
			{"missing", `{}`, "profile alias must not be blank"},
			{"null", `{"alias":null}`, "profile alias must not be blank"},
		} {
			t.Run(endpoint.method+"/"+tc.name, func(t *testing.T) {
				body, err := json.Marshal(json.RawMessage(tc.payload))
				if err != nil {
					t.Fatal(err)
				}
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Set("validatedToken", &jwthandling.ParticipantUserClaims{})
				c.Request = httptest.NewRequest(endpoint.method, "/v1/user/profiles", strings.NewReader(string(body)))
				endpoint.handler(c)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want 400; body: %s", w.Code, w.Body.String())
				}
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if response["error"] != tc.message {
					t.Fatalf("unexpected validation error: %v", response)
				}
			})
		}
	}
}

func TestProfileRequestsIgnoreServerOwnedFields(t *testing.T) {
	for _, req := range []any{&createProfileRequest{}, &updateProfileRequest{}} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/user/profiles", strings.NewReader(`{"id":"507f1f77bcf86cd799439011","alias":"Test","avatarID":"default","consentConfirmedAt":1700000000000,"mainProfile":"not a boolean","createdAt":"not a timestamp"}`))
		if err := c.ShouldBindJSON(req); err != nil {
			t.Fatalf("legacy extra fields rejected: %v", err)
		}
		switch v := req.(type) {
		case *createProfileRequest:
			if v.Alias != "Test" || v.AvatarID != "default" || v.ConsentConfirmedAt != 1700000000000 {
				t.Fatalf("wrong input: %+v", v)
			}
		case *updateProfileRequest:
			if v.ID.Hex() != "507f1f77bcf86cd799439011" || v.Alias != "Test" || v.AvatarID != "default" || v.ConsentConfirmedAt != 1700000000000 {
				t.Fatalf("wrong input: %+v", v)
			}
		}
	}
}

func TestCreateProfileRequiresConsent(t *testing.T) {
	for _, payload := range []string{`{"alias":"Test"}`, `{"alias":"Test","consentConfirmedAt":0}`, `{"alias":"Test","consentConfirmedAt":-1}`} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("validatedToken", &jwthandling.ParticipantUserClaims{})
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/user/profiles", strings.NewReader(payload))
		(&HttpEndpoints{}).addNewProfileHandl(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}
}

func TestUpdateProfileAliasValidation(t *testing.T) {
	legacy := umUtils.InitNewEmailUser("user@organisation.example", "unused", "en").Profiles[0]
	if err := umUtils.ValidateProfileAlias(legacy.Alias); err == nil {
		t.Fatal("fixture must have a legacy alias exceeding the current limit")
	}
	for _, tc := range []struct {
		name, stored, sent string
		valid              bool
	}{
		{"unchanged generated long alias", legacy.Alias, legacy.Alias, true},
		{"unchanged long Unicode alias", strings.Repeat("😀", 11), strings.Repeat("😀", 11), true},
		{"unchanged valid alias", "Original", "Original", true},
		{"rename legacy alias", legacy.Alias, "Updated", true},
		{"rename to limit", legacy.Alias, strings.Repeat("a", 20), true},
		{"changed long alias", legacy.Alias, legacy.Alias + "x", false},
		{"new long alias", "Original", legacy.Alias, false},
		{"long Unicode alias", "Original", strings.Repeat("😀", 11), false},
		{"empty replacement", legacy.Alias, "", false},
		{"blank replacement", legacy.Alias, "   ", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := updateProfileRequest{Alias: tc.sent}
			err := req.validateAlias(userTypes.Profile{Alias: tc.stored})
			if (err == nil) != tc.valid {
				t.Fatalf("valid = %v, want %v; error: %v", err == nil, tc.valid, err)
			}
		})
	}
}
