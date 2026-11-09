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
	"github.com/gin-gonic/gin"
)

func TestProfileHandlersRejectInvalidAlias(t *testing.T) {
	// No database is configured: invalid aliases must be rejected before DB access.
	h := &HttpEndpoints{}
	for _, endpoint := range []struct {
		method  string
		handler gin.HandlerFunc
	}{
		{http.MethodPost, h.addNewProfileHandl},
		{http.MethodPut, h.updateProfileHandl},
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
