// Package instantintegration is used for marketing to capture people interested in instant integration
package instantintegration

import (
	"fmt"
	"log"
	"mash/pkg/analytics"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type InstantIntegrationInterestRepo interface {
	Insert(email string) error
}

type repo struct {
	db *sqlx.DB
}

func NewInstantIntegrationInterestRepo(db *sqlx.DB) InstantIntegrationInterestRepo {
	return &repo{db}
}

func (r *repo) Insert(email string) error {
	_, err := r.db.Exec(`INSERT INTO instant_integration_requested_access (email, created_at)
		VALUES ($1, NOW())`, email)
	if err != nil {
		return fmt.Errorf("failed to perform insert: %w", err)
	}
	return nil
}

type IIAccessRequest struct {
	Email string `json:"email" binding:"required"`
}

func Insert(logger *zap.Logger, analyticsClient analytics.Client, repo InstantIntegrationInterestRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger := logger

		var req IIAccessRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to parse or validate input"})
			return
		}

		req.Email = strings.TrimSpace(req.Email)

		if !strings.Contains(req.Email, "@") {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
			return
		}

		logger = logger.With(zap.String("email", req.Email))

		err := repo.Insert(req.Email)
		if err != nil {
			logger.Error("failed to add to list", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = analyticsClient.Enqueue(&analytics.Capture{
			DistinctId: req.Email,
			Event:      "requested instant integration access",
		})
		if err != nil {
			logger.Error("analytics failed", zap.Error(err))
		}
		c.Status(http.StatusOK)
	}
}