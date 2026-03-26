package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/port"
	"github.com/modami/user-service/pkg/apperror"
)

type AddressService struct {
	addressRepo port.AddressRepository
	cache       port.CacheService
}

func NewAddressService(
	addressRepo port.AddressRepository,
	cache port.CacheService,
) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
		cache:       cache,
	}
}

func (s *AddressService) AddAddress(ctx context.Context, userID uuid.UUID, req dto.CreateAddressRequest) (*domain.Address, error) {
	count, err := s.addressRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 10 {
		return nil, apperror.ErrAddressLimitReached
	}

	now := time.Now()
	addr := &domain.Address{
		ID:            uuid.New(),
		UserID:        userID,
		Label:         req.Label,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		Ward:          req.Ward,
		District:      req.District,
		Province:      req.Province,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		IsDefault:     req.IsDefault || count == 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.addressRepo.Create(ctx, addr); err != nil {
		return nil, err
	}

	_ = s.cache.DeleteAddresses(ctx, userID)
	return addr, nil
}

func (s *AddressService) ListAddresses(ctx context.Context, userID uuid.UUID) ([]*domain.Address, error) {
	cached, err := s.cache.GetAddresses(ctx, userID)
	if err == nil && cached != nil {
		return cached, nil
	}

	addrs, err := s.addressRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetAddresses(ctx, userID, addrs)
	return addrs, nil
}

func (s *AddressService) UpdateAddress(ctx context.Context, userID, addrID uuid.UUID, req dto.UpdateAddressRequest) (*domain.Address, error) {
	addr, err := s.addressRepo.GetByID(ctx, addrID, userID)
	if err != nil {
		return nil, err
	}

	if req.Label != nil {
		addr.Label = *req.Label
	}
	if req.RecipientName != nil {
		addr.RecipientName = *req.RecipientName
	}
	if req.Phone != nil {
		addr.Phone = *req.Phone
	}
	if req.AddressLine1 != nil {
		addr.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		addr.AddressLine2 = *req.AddressLine2
	}
	if req.Ward != nil {
		addr.Ward = *req.Ward
	}
	if req.District != nil {
		addr.District = *req.District
	}
	if req.Province != nil {
		addr.Province = *req.Province
	}
	if req.PostalCode != nil {
		addr.PostalCode = *req.PostalCode
	}
	if req.Country != nil {
		addr.Country = *req.Country
	}
	if req.IsDefault != nil {
		addr.IsDefault = *req.IsDefault
	}
	addr.UpdatedAt = time.Now()

	if err := s.addressRepo.Update(ctx, addr); err != nil {
		return nil, err
	}

	_ = s.cache.DeleteAddresses(ctx, userID)
	return addr, nil
}

func (s *AddressService) DeleteAddress(ctx context.Context, userID, addrID uuid.UUID) error {
	if err := s.addressRepo.Delete(ctx, addrID, userID); err != nil {
		return err
	}
	_ = s.cache.DeleteAddresses(ctx, userID)
	return nil
}

func (s *AddressService) SetDefault(ctx context.Context, userID, addrID uuid.UUID) error {
	if err := s.addressRepo.SetDefault(ctx, addrID, userID); err != nil {
		return err
	}
	_ = s.cache.DeleteAddresses(ctx, userID)
	return nil
}
