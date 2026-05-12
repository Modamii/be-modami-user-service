package middleware

import (
	"context"

	"be-modami-user-service/pkg/pgxtx"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transaction begins a DB transaction before the handler and commits or rolls it back after.
func Transaction(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx, err := db.Begin(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "cannot begin transaction"})
			return
		}

		c.Request = c.Request.WithContext(pgxtx.Inject(c.Request.Context(), tx))

		defer func() {
			if r := recover(); r != nil {
				_ = tx.Rollback(context.Background())
				panic(r)
			}
		}()

		c.Next()

		if c.Writer.Status() >= 400 || len(c.Errors) > 0 {
			_ = tx.Rollback(c.Request.Context())
			return
		}

		if err := tx.Commit(c.Request.Context()); err != nil {
			_ = tx.Rollback(c.Request.Context())
			c.JSON(500, gin.H{"error": "commit failed"})
		}
	}
}
