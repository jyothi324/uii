package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
)

const TraceIDKey = "traceId"

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	
		traceId := c.GetHeader("x-trace-id")
		if traceId == "" {
			traceId = uuid.New().String()
		}

		c.Set("traceId", traceId)
		c.Header("X-Trace-Id", traceId)

		c.Next()
	}
}



func GetTraceID(c *gin.Context) string {
	if v, ok := c.Get(TraceIDKey); ok {
		return v.(string)
	}
	return "unknown-trace"
}
