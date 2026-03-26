package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
	"github.com/modami/user-service/pkg/validator"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// CreateReview godoc
// @Summary      Create a review
// @Description  Create a review for a user (as a buyer)
// @Tags         Reviews
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Reviewee user ID (UUID)"
// @Param        body  body      dto.CreateReviewRequest  true  "Review details"
// @Success      201   {object}  dto.ReviewResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id}/reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	revieweeIDStr := c.Param("id")
	revieweeID, err := uuid.Parse(revieweeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	review, err := h.reviewService.CreateReview(
		c.Request.Context(),
		reviewerID, revieweeID, orderID,
		req.Rating, req.Comment,
		domain.ReviewRoleBuyer,
		req.IsAnonymous,
	)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toReviewResponse(review))
}

// ListReviews godoc
// @Summary      List reviews
// @Description  Returns a paginated list of reviews for a user
// @Tags         Reviews
// @Produce      json
// @Param        id      path      string  true   "User ID (UUID)"
// @Param        limit   query     int     false  "Results per page (max 100)"  default(20)
// @Param        cursor  query     string  false  "Pagination cursor"
// @Success      200     {object}  dto.ReviewListResponse
// @Failure      400     {object}  ErrorResponse
// @Router       /users/{id}/reviews [get]
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	idStr := c.Param("id")
	revieweeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor := c.Query("cursor")

	reviews, nextCursor, err := h.reviewService.ListReviews(c.Request.Context(), revieweeID, limit, cursor)
	if err != nil {
		handleError(c, err)
		return
	}

	items := make([]dto.ReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		items = append(items, toReviewResponse(r))
	}

	c.JSON(http.StatusOK, dto.ReviewListResponse{
		Reviews: items,
		Cursor:  nextCursor,
	})
}

// GetRatingSummary godoc
// @Summary      Get rating summary
// @Description  Returns rating statistics for a user
// @Tags         Reviews
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  dto.RatingSummaryResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /users/{id}/reviews/summary [get]
func (h *ReviewHandler) GetRatingSummary(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	summary, err := h.reviewService.GetRatingSummary(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.RatingSummaryResponse{
		UserID:       summary.UserID.String(),
		AvgRating:    summary.AvgRating,
		TotalReviews: summary.TotalReviews,
		Count1:       summary.Count1,
		Count2:       summary.Count2,
		Count3:       summary.Count3,
		Count4:       summary.Count4,
		Count5:       summary.Count5,
	})
}

func toReviewResponse(r *domain.Review) dto.ReviewResponse {
	resp := dto.ReviewResponse{
		ID:          r.ID.String(),
		RevieweeID:  r.RevieweeID.String(),
		OrderID:     r.OrderID.String(),
		Rating:      r.Rating,
		Comment:     r.Comment,
		Role:        string(r.Role),
		IsAnonymous: r.IsAnonymous,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if !r.IsAnonymous {
		resp.ReviewerID = r.ReviewerID.String()
	}
	return resp
}
