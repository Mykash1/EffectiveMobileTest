package repository

import (
	"context"

	"EffectiveMobileTest/subscriptions-service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id,
			service_name,
			price,
			user_id,
			start_date,
    		end_date
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	)

	return err
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id,
               start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
	`

	var sub model.Subscription

	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM subscriptions WHERE id = $1`,
		id,
	)

	return err
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id,
               start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var sub model.Subscription

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetForCalculation(
	ctx context.Context,
	userID string,
	serviceName string,
) ([]model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id,
               start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE ($1 = '' OR user_id::text = $1)
		AND ($2 = '' OR service_name = $2)
	`

	rows, err := r.db.Query(ctx, query, userID, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var sub model.Subscription

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}
