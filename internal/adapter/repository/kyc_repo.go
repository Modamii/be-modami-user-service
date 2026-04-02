package repository

import (
	"context"
	"time"

	"be-modami-user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type kycRepo struct {
	db *pgxpool.Pool
}

func NewKYCRepository(db *pgxpool.Pool) *kycRepo {
	return &kycRepo{db: db}
}

func (r *kycRepo) Create(ctx context.Context, doc *domain.KYCDocument) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		INSERT INTO kyc_documents (id, user_id, doc_type, doc_url, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		doc.ID, doc.UserID, doc.DocType, doc.DocURL, doc.Status,
		doc.CreatedAt, doc.UpdatedAt,
	)
	return err
}

func (r *kycRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.KYCDocument, error) {
	rows, err := dbFromCtx(ctx, r.db).Query(ctx, `
		SELECT id, user_id, doc_type, doc_url, status, reason, reviewed_by, reviewed_at, created_at, updated_at
		FROM kyc_documents WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.KYCDocument
	for rows.Next() {
		d := &domain.KYCDocument{}
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.DocType, &d.DocURL, &d.Status,
			&d.Reason, &d.ReviewedBy, &d.ReviewedAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (r *kycRepo) UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.KYCDocStatus, reason string, reviewedBy uuid.UUID) error {
	now := time.Now()
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		UPDATE kyc_documents SET status=$1, reason=$2, reviewed_by=$3, reviewed_at=$4, updated_at=$4
		WHERE user_id=$5`,
		status, reason, reviewedBy, now, userID,
	)
	return err
}

type sellerRepo struct {
	db *pgxpool.Pool
}

func NewSellerProfileRepository(db *pgxpool.Pool) *sellerRepo {
	return &sellerRepo{db: db}
}

func (r *sellerRepo) Create(ctx context.Context, profile *domain.SellerProfile) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		INSERT INTO seller_profiles (id, user_id, shop_name, shop_slug, shop_description, shop_logo_url,
			shop_banner_url, business_type, tax_id, bank_account, bank_name, kyc_status, avg_rating,
			total_reviews, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		profile.ID, profile.UserID, profile.ShopName, profile.ShopSlug,
		profile.ShopDescription, profile.ShopLogoURL, profile.ShopBannerURL,
		profile.BusinessType, profile.TaxID, profile.BankAccount, profile.BankName,
		profile.KYCStatus, profile.AvgRating, profile.TotalReviews,
		profile.CreatedAt, profile.UpdatedAt,
	)
	return err
}

func (r *sellerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.SellerProfile, error) {
	row := dbFromCtx(ctx, r.db).QueryRow(ctx, `
		SELECT id, user_id, shop_name, shop_slug, shop_description, shop_logo_url, shop_banner_url,
			business_type, tax_id, bank_account, bank_name, kyc_status, kyc_verified_at,
			avg_rating, total_reviews, created_at, updated_at
		FROM seller_profiles WHERE user_id=$1`, userID)

	p := &domain.SellerProfile{}
	err := row.Scan(
		&p.ID, &p.UserID, &p.ShopName, &p.ShopSlug, &p.ShopDescription,
		&p.ShopLogoURL, &p.ShopBannerURL, &p.BusinessType, &p.TaxID,
		&p.BankAccount, &p.BankName, &p.KYCStatus, &p.KYCVerifiedAt,
		&p.AvgRating, &p.TotalReviews, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r *sellerRepo) Update(ctx context.Context, profile *domain.SellerProfile) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		UPDATE seller_profiles SET shop_name=$1, shop_description=$2, shop_logo_url=$3, shop_banner_url=$4,
			tax_id=$5, bank_account=$6, bank_name=$7, updated_at=$8
		WHERE user_id=$9`,
		profile.ShopName, profile.ShopDescription, profile.ShopLogoURL, profile.ShopBannerURL,
		profile.TaxID, profile.BankAccount, profile.BankName, time.Now(), profile.UserID,
	)
	return err
}

func (r *sellerRepo) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status domain.KYCStatus, verifiedAt *time.Time) error {
	_, err := dbFromCtx(ctx, r.db).Exec(ctx, `
		UPDATE seller_profiles SET kyc_status=$1, kyc_verified_at=$2, updated_at=$3 WHERE user_id=$4`,
		status, verifiedAt, time.Now(), userID,
	)
	return err
}
