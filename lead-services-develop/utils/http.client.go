package utils

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuliporg/agent-management-api/middleware"
)

func NewTracedRequest(
	c *gin.Context,
	method string,
	url string,
	body []byte,
) (*http.Request, error) {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	traceId := middleware.GetTraceID(c)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-trace-id", traceId)

	return req, nil
}
