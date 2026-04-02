package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/service"
)

type UserRegisteredHandler struct {
	userService *service.UserService
}

func NewUserRegisteredHandler(userService *service.UserService) *UserRegisteredHandler {
	return &UserRegisteredHandler{userService: userService}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, payload []byte) error {
	var event domain.AuthUserCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal AuthUserCreatedEvent: %w", err)
	}
	return h.userService.CreateFromEvent(ctx, &event)
}
