package httputil

import "github.com/gin-gonic/gin"

// NewError example
func NewError(ctx *gin.Context, status int, err error) {
	e := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	ctx.JSON(status, e)
}

// HTTPError struct
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"bad request"`
}
