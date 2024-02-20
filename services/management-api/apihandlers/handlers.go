package apihandlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlerTest() {
	// ...
	slog.Info("HandlerTest called")
}

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	// db connections
}

func NewHTTPHandler(
// db connections
) *HttpEndpoints {

	return &HttpEndpoints{}
}
