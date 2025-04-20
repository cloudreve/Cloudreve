package middleware

import (
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs/dbfs"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/cloudreve/Cloudreve/v4/routers/controllers"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

// UrisService is a wrapper for service supports batch file operations
type UrisService interface {
	GetUris() []string
}

// ValidateBatchFileCount validates if the batch file count is within the limit
func ValidateBatchFileCount(dep dependency.Dep, ctxKey interface{}) gin.HandlerFunc {
	settings := dep.SettingProvider()
	return func(c *gin.Context) {
		uris := controllers.ParametersFromContext[UrisService](c, ctxKey)
		limit := settings.MaxBatchedFile(c)
		if len((uris).GetUris()) > limit {
			c.JSON(200, serializer.ErrWithDetails(
				c,
				serializer.CodeTooManyUris,
				fmt.Sprintf("Maximum allowed batch size: %d", limit),
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ContextHint parses the context hint header and set it to context
func ContextHint() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(dbfs.ContextHintHeader) != "" {
			util.WithValue(c, dbfs.ContextHintCtxKey{}, uuid.FromStringOrNil(c.GetHeader(dbfs.ContextHintHeader)))
		}

		c.Next()
	}
}
