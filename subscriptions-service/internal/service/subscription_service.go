package service

import (
	"context"
	"time"

	"EffectiveMobileTest/subscriptions-service/internal/dto"
	"EffectiveMobileTest/subscriptions-service/internal/model"
	"EffectiveMobileTest/subscriptions-service/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func parseMonth(date string) (time.Time, error) {
	return time.Parse("01-2006", date)
}

func (s *SubscriptionService) Create(
	ctx context.Context,
	req dto.CreateSubscriptionRequest,
) error {
	startDate, err := parseMonth(req.StartDate)
	if err != nil {
		return err
	}

	var endDate *time.Time

	if req.EndDate != "" {
		parsed, err := parseMonth(req.EndDate)
		if err != nil {
			return err
		}

		endDate = &parsed
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return err
	}

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, parsedID)
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, parsedID)
}

func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	return s.repo.List(ctx)
}

func (s *SubscriptionService) CalculateTotal(
	ctx context.Context,
	from string,
	to string,
	userID string,
	serviceName string,
) (int, error) {

	fromDate, err := parseMonth(from)
	if err != nil {
		return 0, err
	}

	toDate, err := parseMonth(to)
	if err != nil {
		return 0, err
	}

	subscriptions, err := s.repo.GetForCalculation(ctx, userID, serviceName)
	if err != nil {
		return 0, err
	}

	total := 0

	for _, sub := range subscriptions {
		//	subscriptions := toDate

		if sub.EndDate != nil {
			subscriptionEnd := *sub.EndDate

			if subscriptionEnd.Before(toDate) {
				continue
			}
		}

		current := sub.StartDate

		for !current.After(toDate) {
			if !current.Before(fromDate) && !current.After(toDate) {
				total += sub.Price
			}
			current = current.AddDate(0, 1, 0)
		}
	}
	return total, nil
}
