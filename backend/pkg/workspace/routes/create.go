package routes

import (
	"mash/pkg/auth"
	service_workspace "mash/pkg/workspace/service"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"

	"mash/pkg/codebase/access"
	db_codebase "mash/pkg/codebase/db"
)

type CreateRequest struct {
	CodebaseID string `json:"codebase_id" binding:"required"`
	Name       string `json:"name"`

	// ChangeID is a commit checksum
	ChangeID string `json:"change_id"` // change_id and revert_change_id are mutually exclusive

	// RevertChangeID is a commit checksum
	RevertChangeID string `json:"revert_change_id"` //
}

func Create(logger *zap.Logger, workspaceService service_workspace.Service, codebaseUserRepo db_codebase.CodebaseUserRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		userID, err := auth.UserID(c.Request.Context())
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("failed to parse input", zap.Error(err))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		req.CodebaseID = strings.TrimSpace(req.CodebaseID)

		if !access.UserHasAccessToCodebase(codebaseUserRepo, userID, req.CodebaseID) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ws, err := workspaceService.Create(service_workspace.CreateWorkspaceRequest{
			UserID:         userID,
			CodebaseID:     req.CodebaseID,
			Name:           req.Name,
			RevertChangeID: req.RevertChangeID,
			ChangeID:       req.ChangeID,
		})

		if err != nil {
			logger.Error("failed to create workspace", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, ws)
	}
}