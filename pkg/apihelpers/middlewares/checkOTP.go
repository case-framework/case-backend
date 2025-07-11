package middlewares

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"github.com/gin-gonic/gin"
)

type OTPConfig struct {
	Route  string              `json:"route" yaml:"route"`
	Method string              `json:"method" yaml:"method"`
	Exact  bool                `json:"exact" yaml:"exact"`
	MaxAge time.Duration       `json:"maxAge" yaml:"max_age"`
	Types  []userTypes.OTPType `json:"types" yaml:"types"`
}

func CheckOTP(otpConf []OTPConfig, tokenSignKey string, globalInfosDBService *globalinfosDB.GlobalInfosDBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.Request.URL.Path
		method := c.Request.Method

		conf := getOTPConfigForRoute(route, method, otpConf)
		if conf == nil {
			// no OTP is required for this route
			c.Next()
			return
		}

		extractAndValidateParticipantJWT(c, tokenSignKey, globalInfosDBService)

		tokenValue, ok := c.Get("validatedToken")
		if !ok {
			slog.Warn("validatedToken not found in context")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing access token"})
			return
		}
		parsedToken := tokenValue.(*jwthandling.ParticipantUserClaims)

		if parsedToken.LastOTPProvided == nil {
			slog.Warn("no OTP provided", slog.String("instanceID", parsedToken.InstanceID), slog.String("userID", parsedToken.Subject), slog.String("route", route))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing OTP"})
			return
		}

		for _, otpType := range conf.Types {
			lastOTPProvidedValue, ok := parsedToken.LastOTPProvided[string(otpType)]
			if !ok {
				continue
			}

			if lastOTPProvidedValue >= time.Now().Unix()-int64(conf.MaxAge.Seconds()) {
				// OTP is valid
				c.Next()
				return
			}
		}
		slog.Warn("no or expired OTP for required types", slog.String("instanceID", parsedToken.InstanceID), slog.String("userID", parsedToken.Subject), slog.String("route", route))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid OTP"})
	}
}

func getOTPConfigForRoute(route string, method string, otpConf []OTPConfig) *OTPConfig {
	var foundConfig *OTPConfig

	for _, conf := range otpConf {
		if conf.Method != "" && conf.Method != method {
			continue
		}

		if conf.Exact && conf.Route == route {
			return &conf
		}

		if strings.HasPrefix(route, conf.Route) {
			// if route from config is longer than the previous found route:
			if foundConfig == nil || len(conf.Route) > len(foundConfig.Route) {
				foundConfig = &conf
			}
		}
	}
	return foundConfig
}
