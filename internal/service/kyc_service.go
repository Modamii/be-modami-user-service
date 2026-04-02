package service

import (
	"context"
	"encoding/json"
	"time"

	apperror "be-modami-user-service/internal/apperror"
	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/dto"
	"be-modami-user-service/internal/port"

	"github.com/google/uuid"
)

type KYCService struct {
	kycRepo    port.KYCRepository
	sellerRepo port.SellerProfileRepository
	userRepo   port.UserRepository
	cache      port.CacheService
	txManager  port.TxManager
	outboxRepo port.OutboxRepository
	topic      string
}

func NewKYCService(
	kycRepo port.KYCRepository,
	sellerRepo port.SellerProfileRepository,
	userRepo port.UserRepository,
	cache port.CacheService,
	txManager port.TxManager,
	outboxRepo port.OutboxRepository,
	topic string,
) *KYCService {
	return &KYCService{
		kycRepo:    kycRepo,
		sellerRepo: sellerRepo,
		userRepo:   userRepo,
		cache:      cache,
		txManager:  txManager,
		outboxRepo: outboxRepo,
		topic:      topic,
	}
}

func (s *KYCService) SubmitKYC(ctx context.Context, userID uuid.UUID, req dto.SubmitKYCRequest) error {
	profile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return apperror.ErrSellerNotFound
	}
	if profile.KYCStatus == domain.KYCStatusApproved {
		return apperror.ErrInvalidKYCState
	}

	now := time.Now()
	for _, doc := range req.Documents {
		kycDoc := &domain.KYCDocument{
			ID:        uuid.New(),
			UserID:    userID,
			DocType:   domain.KYCDocType(doc.DocType),
			DocURL:    doc.DocURL,
			Status:    domain.KYCDocStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.kycRepo.Create(ctx, kycDoc); err != nil {
			return err
		}
	}

	if err := s.sellerRepo.UpdateKYCStatus(ctx, userID, domain.KYCStatusPending, nil); err != nil {
		return err
	}

	_ = s.cache.DeleteKYCStatus(ctx, userID)
	_ = s.cache.DeleteSellerProfile(ctx, userID)
	return nil
}

func (s *KYCService) GetKYCStatus(ctx context.Context, userID uuid.UUID) (domain.KYCStatus, error) {
	cached, err := s.cache.GetKYCStatus(ctx, userID)
	if err == nil && cached != "" {
		return cached, nil
	}

	profile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if profile == nil {
		return domain.KYCStatusNone, nil
	}

	_ = s.cache.SetKYCStatus(ctx, userID, profile.KYCStatus)
	return profile.KYCStatus, nil
}

func (s *KYCService) ApproveKYC(ctx context.Context, userID, adminID uuid.UUID) error {
	profile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return apperror.ErrSellerNotFound
	}
	if profile.KYCStatus != domain.KYCStatusPending {
		return apperror.ErrInvalidKYCState
	}

	// Upgrade user role to seller
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	oldRole := user.Role

	now := time.Now()

	return s.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.kycRepo.UpdateStatus(ctx, userID, domain.KYCDocStatusApproved, "", adminID); err != nil {
			return err
		}
		if err := s.sellerRepo.UpdateKYCStatus(ctx, userID, domain.KYCStatusApproved, &now); err != nil {
			return err
		}
		if err := s.userRepo.UpdateRole(ctx, userID, domain.UserRoleSeller); err != nil {
			return err
		}

		_ = s.cache.DeleteKYCStatus(ctx, userID)
		_ = s.cache.DeleteSellerProfile(ctx, userID)
		_ = s.cache.DeleteProfile(ctx, userID)

		payload, _ := json.Marshal(&domain.UserRoleUpgradedEvent{
			UserID:  userID,
			OldRole: oldRole,
			NewRole: domain.UserRoleSeller,
		})
		return s.outboxRepo.Create(ctx, s.topic, userID.String(), payload)
	})
}

func (s *KYCService) RejectKYC(ctx context.Context, userID, adminID uuid.UUID, reason string) error {
	profile, err := s.sellerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return apperror.ErrSellerNotFound
	}
	if profile.KYCStatus != domain.KYCStatusPending {
		return apperror.ErrInvalidKYCState
	}

	if err := s.kycRepo.UpdateStatus(ctx, userID, domain.KYCDocStatusRejected, reason, adminID); err != nil {
		return err
	}

	if err := s.sellerRepo.UpdateKYCStatus(ctx, userID, domain.KYCStatusRejected, nil); err != nil {
		return err
	}

	_ = s.cache.DeleteKYCStatus(ctx, userID)
	_ = s.cache.DeleteSellerProfile(ctx, userID)
	return nil
}
