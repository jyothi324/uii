package utils

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
)

func Log(c *gin.Context, format string, args ...any) {
	traceId := middleware.GetTraceID(c)
	log.Printf("[traceId=%s] "+format, append([]any{traceId}, args...)...)
}
