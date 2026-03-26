package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/pkg/apperror"
)

type addressRepo struct {
	db *pgxpool.Pool
}

func NewAddressRepository(db *pgxpool.Pool) *addressRepo {
	return &addressRepo{db: db}
}

func (r *addressRepo) Create(ctx context.Context, addr *domain.Address) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO addresses (id, user_id, label, recipient_name, phone, address_line_1, address_line_2,
			ward, district, province, postal_code, country, is_default, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		addr.ID, addr.UserID, addr.Label, addr.RecipientName, addr.Phone,
		addr.AddressLine1, addr.AddressLine2, addr.Ward, addr.District, addr.Province,
		addr.PostalCode, addr.Country, addr.IsDefault, addr.CreatedAt, addr.UpdatedAt,
	)
	return err
}

func (r *addressRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Address, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, label, recipient_name, phone, address_line_1, address_line_2,
			ward, district, province, postal_code, country, is_default, created_at, updated_at
		FROM addresses WHERE id = $1 AND user_id = $2`, id, userID)
	return r.scanAddress(row)
}

func (r *addressRepo) scanAddress(row pgx.Row) (*domain.Address, error) {
	a := &domain.Address{}
	err := row.Scan(
		&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
		&a.AddressLine1, &a.AddressLine2, &a.Ward, &a.District, &a.Province,
		&a.PostalCode, &a.Country, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrAddressNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *addressRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Address, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, label, recipient_name, phone, address_line_1, address_line_2,
			ward, district, province, postal_code, country, is_default, created_at, updated_at
		FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []*domain.Address
	for rows.Next() {
		a := &domain.Address{}
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
			&a.AddressLine1, &a.AddressLine2, &a.Ward, &a.District, &a.Province,
			&a.PostalCode, &a.Country, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	return addrs, rows.Err()
}

func (r *addressRepo) Update(ctx context.Context, addr *domain.Address) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE addresses SET label=$1, recipient_name=$2, phone=$3, address_line_1=$4, address_line_2=$5,
			ward=$6, district=$7, province=$8, postal_code=$9, country=$10, is_default=$11, updated_at=$12
		WHERE id=$13 AND user_id=$14`,
		addr.Label, addr.RecipientName, addr.Phone, addr.AddressLine1, addr.AddressLine2,
		addr.Ward, addr.District, addr.Province, addr.PostalCode, addr.Country,
		addr.IsDefault, time.Now(), addr.ID, addr.UserID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperror.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM addresses WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperror.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepo) SetDefault(ctx context.Context, id, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE addresses SET is_default=false, updated_at=$1 WHERE user_id=$2 AND is_default=true`,
		time.Now(), userID,
	)
	if err != nil {
		return err
	}

	ct, err := tx.Exec(ctx,
		`UPDATE addresses SET is_default=true, updated_at=$1 WHERE id=$2 AND user_id=$3`,
		time.Now(), id, userID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperror.ErrAddressNotFound
	}

	return tx.Commit(ctx)
}

func (r *addressRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM addresses WHERE user_id=$1`, userID).Scan(&count)
	return count, err
}
