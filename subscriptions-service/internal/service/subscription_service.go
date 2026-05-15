package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"EffectiveMobileTest/internal/dto"
	"EffectiveMobileTest/internal/model"
	"EffectiveMobileTest/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo   *repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo *repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

func parseMonth(date string) (time.Time, error) {
	t, err := time.Parse("01-2006", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format %q, expected MM-YYYY", date)
	}
	return t, nil
}

func (s *SubscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
	startDate, err := parseMonth(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	var endDate *time.Time
	if req.EndDate != "" {
		parsed, err := parseMonth(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		if parsed.Before(startDate) {
			return nil, errors.New("end_date must be after or equal to start_date")
		}
		endDate = &parsed
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info("subscription created", "id", sub.ID, "service_name", sub.ServiceName, "user_id", sub.UserID)
	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", ErrNotFound)
	}

	sub, err := s.repo.GetByID(ctx, parsedID)
	if err != nil {
		return nil, ErrNotFound
	}

	s.logger.Info("subscription retrieved", "id", sub.ID)
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", ErrNotFound)
	}

	existing, err := s.repo.GetByID(ctx, parsedID)
	if err != nil {
		return nil, ErrNotFound
	}

	if req.ServiceName != nil {
		existing.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		if *req.Price < 1 {
			return nil, errors.New("price must be at least 1")
		}
		existing.Price = *req.Price
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			existing.EndDate = nil
		} else {
			parsed, err := parseMonth(*req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end_date: %w", err)
			}
			if parsed.Before(existing.StartDate) {
				return nil, errors.New("end_date must be after or equal to start_date")
			}
			existing.EndDate = &parsed
		}
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, ErrNotFound
	}

	s.logger.Info("subscription updated", "id", existing.ID)
	return existing, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ErrNotFound
	}

	if err := s.repo.Delete(ctx, parsedID); err != nil {
		return ErrNotFound
	}

	s.logger.Info("subscription deleted", "id", parsedID)
	return nil
}

type ListResult struct {
	Subscriptions []model.Subscription `json:"subscriptions"`
	Total         int                  `json:"total"`
	Page          int                  `json:"page"`
	PerPage       int                  `json:"per_page"`
}

func (s *SubscriptionService) List(ctx context.Context, page int, perPage int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	subs, err := s.repo.List(ctx, perPage, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Info("subscriptions listed", "page", page, "per_page", perPage, "total", total)
	return &ListResult{
		Subscriptions: subs,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
	}, nil
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, from string, to string, userID string, serviceName string) (int, error) {
	if from == "" || to == "" {
		return 0, errors.New("from and to parameters are required")
	}

	fromDate, err := parseMonth(from)
	if err != nil {
		return 0, fmt.Errorf("invalid from: %w", err)
	}

	toDate, err := parseMonth(to)
	if err != nil {
		return 0, fmt.Errorf("invalid to: %w", err)
	}

	if fromDate.After(toDate) {
		return 0, errors.New("from date must be before or equal to to date")
	}

	subscriptions, err := s.repo.GetForCalculation(ctx, userID, serviceName)
	if err != nil {
		return 0, err
	}

	total := 0

	for _, sub := range subscriptions {
		months, err := countMonthsInIntersection(sub.StartDate, sub.EndDate, fromDate, toDate)
		if err != nil {
			return 0, fmt.Errorf("error calculating for subscription %s: %w", sub.ID, err)
		}
		total += months * sub.Price
	}

	s.logger.Info("total calculated", "from", from, "to", to, "user_id", userID, "service_name", serviceName, "total", total)
	return total, nil
}

// countMonthsInIntersection counts how many full months fall within the intersection
// of the subscription period [subStart, subEnd] and the query period [fromDate, toDate].
// If subEnd is nil, the subscription is considered active indefinitely.
func countMonthsInIntersection(subStart time.Time, subEnd *time.Time, fromDate, toDate time.Time) (int, error) {
	// Determine the effective start: later of subscription start and query from
	effectiveStart := subStart
	if fromDate.After(effectiveStart) {
		effectiveStart = fromDate
	}

	// Determine the effective end: earlier of subscription end and query to
	var effectiveEnd time.Time
	if subEnd != nil {
		effectiveEnd = *subEnd
		if toDate.Before(effectiveEnd) {
			effectiveEnd = toDate
		}
	} else {
		effectiveEnd = toDate
	}

	// No intersection if effective start is after effective end
	if effectiveStart.After(effectiveEnd) {
		return 0, nil
	}

	// Count months from effectiveStart to effectiveEnd inclusive
	count := 0
	current := effectiveStart
	for !current.After(effectiveEnd) {
		count++
		current = current.AddDate(0, 1, 0)
	}

	return count, nil
}
