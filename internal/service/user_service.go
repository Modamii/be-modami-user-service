package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/dto"
	"be-modami-user-service/internal/port"
	apperror "be-modami-user-service/pkg/apperror"
	"be-modami-user-service/pkg/pagination"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo   port.UserRepository
	txManager  port.TxManager
	outboxRepo port.OutboxRepository
}

func NewUserService(
	userRepo port.UserRepository,
	txManager port.TxManager,
	outboxRepo port.OutboxRepository,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		txManager:  txManager,
		outboxRepo: outboxRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserService) GetMyProfile(ctx context.Context, keycloakID string) (*domain.User, error) {
	return s.userRepo.GetByKeycloakID(ctx, keycloakID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	changedFields := map[string]interface{}{}

	if req.FullName != nil {
		user.FullName = *req.FullName
		changedFields["full_name"] = *req.FullName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
		changedFields["phone"] = *req.Phone
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
		changedFields["bio"] = *req.Bio
	}
	if req.Gender != nil {
		user.Gender = domain.GenderType(*req.Gender)
		changedFields["gender"] = *req.Gender
	}
	if req.DateOfBirth != nil {
		t, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, err
		}
		user.DateOfBirth = &t
		changedFields["date_of_birth"] = *req.DateOfBirth
	}

	if err := s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Update(ctx, user); err != nil {
			return err
		}
		if len(changedFields) > 0 {
			payload, _ := json.Marshal(&domain.UserUpdatedEvent{UserID: userID, ChangedFields: changedFields})
			return s.outboxRepo.Create(ctx, domain.OutboxAggregateUser, userID.String(), domain.OutboxEventUserUpdated, payload)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.AvatarURL = avatarURL
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdateCover(ctx context.Context, userID uuid.UUID, coverURL string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.CoverURL = coverURL
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) DeactivateAccount(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.SoftDelete(ctx, userID)
}

func (s *UserService) SearchUsers(ctx context.Context, query string, limit int, cursorStr string) ([]*domain.User, string, error) {
	if limit <= 0 {
		limit = 20
	}

	cursor, err := pagination.DecodeCursor(cursorStr)
	if err != nil {
		return nil, "", err
	}

	users, err := s.userRepo.Search(ctx, query, limit+1, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(users) > limit {
		nextCursor = pagination.EncodeCursor(users[limit-1].CreatedAt)
		users = users[:limit]
	}

	return users, nextCursor, nil
}

func (s *UserService) UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus, reason string) error {
	return s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
			return err
		}
		if status == domain.UserStatusSuspended {
			payload, _ := json.Marshal(&domain.UserSuspendedEvent{UserID: userID, Reason: reason, SuspendedAt: time.Now()})
			return s.outboxRepo.Create(ctx, domain.OutboxAggregateUser, userID.String(), domain.OutboxEventUserSuspended, payload)
		}
		return nil
	})
}

func (s *UserService) MarkEmailVerified(ctx context.Context, keycloakID string) error {
	user, err := s.userRepo.GetByKeycloakID(ctx, keycloakID)
	if err != nil {
		return err
	}
	user.EmailVerified = true
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) SoftDeleteByKeycloakID(ctx context.Context, keycloakID string) error {
	user, err := s.userRepo.GetByKeycloakID(ctx, keycloakID)
	if err != nil {
		if err == apperror.ErrNotFound {
			return nil
		}
		return err
	}
	return s.userRepo.SoftDelete(ctx, user.ID)
}

func (s *UserService) CreateFromEvent(ctx context.Context, event *domain.AuthUserCreatedEvent) error {
	existing, err := s.userRepo.GetByKeycloakID(ctx, event.UserID)
	if err != nil && err != apperror.ErrNotFound {
		return err
	}
	if existing != nil {
		return nil
	}

	now := time.Now()
	user := &domain.User{
		ID:            uuid.New(),
		KeycloakID:    event.UserID,
		Email:         event.Email,
		UserName:      event.Username,
		FullName:      strings.TrimSpace(event.FirstName + " " + event.LastName),
		Role:          domain.UserRoleBuyer,
		Status:        domain.UserStatusActive,
		EmailVerified: false,
		TrustScore:    0,
		Gender:        domain.GenderUndisclosed,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	outboxPayload, err := json.Marshal(&domain.UserProfileCreatedEvent{
		UserID:     user.ID,
		KeycloakID: user.KeycloakID,
		Role:       user.Role,
		Status:     user.Status,
	})
	if err != nil {
		return err
	}

	return s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}
		return s.outboxRepo.Create(ctx, domain.OutboxAggregateUser, user.ID.String(), domain.OutboxEventUserProfileCreated, outboxPayload)
	})
}

func (s *UserService) SyncFromAuthUpdate(ctx context.Context, event *domain.AuthUserUpdatedEvent) error {
	user, err := s.userRepo.GetByKeycloakID(ctx, event.UserID)
	if err != nil {
		return err
	}

	if event.Email != nil {
		user.Email = *event.Email
	}
	if event.FirstName != nil || event.LastName != nil {
		firstName := event.FirstName
		lastName := event.LastName
		if firstName == nil {
			firstName = &user.FullName
		}
		if lastName == nil {
			lastName = new(string)
		}
		user.FullName = strings.TrimSpace(*firstName + " " + *lastName)
	}

	return s.userRepo.Update(ctx, user)
}

// SyncFromKeycloakCDC dispatches a Debezium CDC event for the Keycloak
func (s *UserService) SyncFromKeycloakCDC(ctx context.Context, event *domain.KeycloakCDCEvent) error {
	switch event.Op {
	case domain.KeycloakCDCOpCreate, domain.KeycloakCDCOpSnapshot:
		if event.After == nil || event.After.IsServiceAccount() {
			return nil
		}
		return s.createOrSkipFromCDC(ctx, event.After)
	case domain.KeycloakCDCOpUpdate:
		if event.After == nil || event.After.IsServiceAccount() {
			return nil
		}
		return s.updateFromCDC(ctx, event)
	case domain.KeycloakCDCOpDelete:
		if event.Before == nil {
			return nil
		}
		return s.SoftDeleteByKeycloakID(ctx, event.Before.ID)
	}
	return nil
}

func (s *UserService) createOrSkipFromCDC(ctx context.Context, entity *domain.KeycloakUserEntity) error {
	authEvt := &domain.AuthUserCreatedEvent{
		UserID:    entity.ID,
		Timestamp: time.Now(),
	}
	if entity.Email != nil && *entity.Email != "" {
		authEvt.Email = *entity.Email
	}
	if entity.Username != nil {
		authEvt.Username = *entity.Username
	}
	if entity.FirstName != nil {
		authEvt.FirstName = *entity.FirstName
	}
	if entity.LastName != nil {
		authEvt.LastName = *entity.LastName
	}
	if err := s.CreateFromEvent(ctx, authEvt); err != nil {
		return err
	}
	if entity.EmailVerified {
		return s.MarkEmailVerified(ctx, entity.ID)
	}
	return nil
}

func (s *UserService) updateFromCDC(ctx context.Context, event *domain.KeycloakCDCEvent) error {
	entity := event.After

	diff := event.DetectChanges()
	if !diff.HasAny() {
		return nil
	}

	user, err := s.userRepo.GetByKeycloakID(ctx, entity.ID)
	if err != nil {
		if err == apperror.ErrNotFound {
			return s.createOrSkipFromCDC(ctx, entity)
		}
		return err
	}

	// Build sync fields starting from the current DB values so that fields not
	// changed in this event are preserved unchanged.
	syncFields := domain.KeycloakSyncFields{
		Email:         user.Email,
		Username:      user.UserName,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
	}

	if diff.Email && entity.Email != nil && *entity.Email != "" {
		syncFields.Email = *entity.Email
	}
	if diff.Username && entity.Username != nil && *entity.Username != "" {
		syncFields.Username = *entity.Username
	}
	if diff.FirstName || diff.LastName {
		firstName := ""
		if entity.FirstName != nil {
			firstName = *entity.FirstName
		}
		lastName := ""
		if entity.LastName != nil {
			lastName = *entity.LastName
		}
		if fullName := strings.TrimSpace(firstName + " " + lastName); fullName != "" {
			user.FullName = fullName
		}
	}
	if diff.EmailVerified {
		syncFields.EmailVerified = entity.EmailVerified
	}
	// Mirror Keycloak enabled flag → active/inactive; never override suspended/banned.
	if diff.Enabled {
		if !entity.Enabled && user.Status == domain.UserStatusActive {
			syncFields.Status = domain.UserStatusInactive
		} else if entity.Enabled && user.Status == domain.UserStatusInactive {
			syncFields.Status = domain.UserStatusActive
		}
	}

	if err := s.userRepo.UpdateKeycloakSyncFields(ctx, user.ID, syncFields); err != nil {
		return err
	}
	// Update profile fields (full_name) if name changed.
	if diff.FirstName || diff.LastName {
		return s.userRepo.Update(ctx, user)
	}
	return nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) GetByKeycloakID(ctx context.Context, keycloakID string) (*domain.User, error) {
	return s.userRepo.GetByKeycloakID(ctx, keycloakID)
}
