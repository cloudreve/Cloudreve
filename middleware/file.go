package middleware

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// ValidateSourceLink validates if the perm source link is a valid redirect link
func ValidateSourceLink() gin.HandlerFunc {
	return func(c *gin.Context) {
		linkID, ok := c.Get("object_id")
		if !ok {
			c.JSON(200, serializer.Err(serializer.CodeFileNotFound, "", nil))
			c.Abort()
			return
		}

		sourceLink, err := model.GetSourceLinkByID(linkID)
		if err != nil || sourceLink.File.ID == 0 || sourceLink.File.Name != c.Param("name") {
			c.JSON(200, serializer.Err(serializer.CodeFileNotFound, "", nil))
			c.Abort()
			return
		}

		sourceLink.Downloaded()
		c.Set("source_link", sourceLink)
		c.Next()
	}
}
