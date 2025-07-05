package debt

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/model"
	"github.com/pvxdv/self_improver/internal/service/finance"
	"github.com/pvxdv/self_improver/internal/storage"
)

type debtManager interface {
	AddDebt(ctx context.Context, debt *model.Debt) (int64, error)
	GetDebt(ctx context.Context, id int64) (model.Debt, error)
	UpdateDebt(ctx context.Context, debt model.Debt) error
	DeleteDebt(ctx context.Context, id int64) error
	GetAmountDebt(ctx context.Context) (*model.AmountDebt, error)
}

type ServiceDebt struct {
	debtManager debtManager
	logger      *zap.SugaredLogger
}

func NewDebtService(debtManager debtManager, logger *zap.SugaredLogger) *ServiceDebt {
	return &ServiceDebt{
		debtManager: debtManager,
		logger:      logger,
	}
}

func (s *ServiceDebt) AddDebt(ctx context.Context, debt *model.Debt) error {
	if debt.Amount <= 0 {
		return finance.ErrNegativeAmount
	}

	if len(debt.Description) > 1000 {
		return finance.ErrDescriptionTooLong
	}

	id, err := s.debtManager.AddDebt(ctx, debt)
	if err != nil {
		s.logger.Errorf("failed to add debt: %v", err)
		return fmt.Errorf("failed to add debt")
	}

	s.logger.Infof("successfully added debt ID: %d, Amount: %d, Description: %s",
		id, debt.Amount, debt.Description)
	return nil
}

func (s *ServiceDebt) GetDebt(ctx context.Context, id int64) (model.Debt, error) {
	if id <= 0 {
		return model.Debt{}, finance.ErrInvalidID
	}

	debt, err := s.debtManager.GetDebt(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrDebtNotFound) {
			s.logger.Warnf("debt not found, ID: %d", id)
			return model.Debt{}, err
		}
		s.logger.Errorf("failed to get debt ID %d: %v", id, err)
		return model.Debt{}, fmt.Errorf("failed to get debt")
	}

	return debt, nil
}

func (s *ServiceDebt) UpdateDebt(ctx context.Context, debt model.Debt) error {
	if debt.ID <= 0 {
		return finance.ErrInvalidID
	}

	if debt.Amount <= 0 {
		return finance.ErrNegativeAmount
	}

	if len(debt.Description) > 1000 {
		return finance.ErrDescriptionTooLong
	}

	err := s.debtManager.UpdateDebt(ctx, debt)
	if err != nil {
		if errors.Is(err, storage.ErrDebtNotFound) {
			s.logger.Warnf("debt not found for update, ID: %d", debt.ID)
			return err
		}
		s.logger.Errorf("failed to update debt ID %d: %v", debt.ID, err)
		return fmt.Errorf("failed to update debt")
	}

	s.logger.Infof("successfully updated debt ID: %d", debt.ID)
	return nil
}

func (s *ServiceDebt) DeleteDebt(ctx context.Context, id int64) error {
	if id <= 0 {
		return finance.ErrInvalidID
	}

	err := s.debtManager.DeleteDebt(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrDebtNotFound) {
			s.logger.Warnf("debt not found for deletion, ID: %d", id)
			return err
		}
		s.logger.Errorf("failed to delete debt ID %d: %v", id, err)
		return fmt.Errorf("failed to delete debt")
	}

	s.logger.Infof("successfully deleted debt ID: %d", id)
	return nil
}

func (s *ServiceDebt) GetAmountDebt(ctx context.Context) (*model.AmountDebt, error) {
	amountDebt, err := s.debtManager.GetAmountDebt(ctx)
	if err != nil {
		s.logger.Errorf("failed to get aggregated debt info: %v", err)
		return nil, fmt.Errorf("failed to get debt information")
	}

	s.logger.Info("successfully retrieved aggregated debt info")
	return amountDebt, nil
}

func (s *ServiceDebt) PayDebt(ctx context.Context, id int64, amount int64) error {
	if id <= 0 {
		return finance.ErrInvalidID
	}

	if amount <= 0 {
		return finance.ErrNegativeAmount
	}

	debt, err := s.debtManager.GetDebt(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrDebtNotFound) {
			s.logger.Warnf("debt not found for payment, ID: %d", id)
			return err
		}
		s.logger.Errorf("failed to get debt for payment, ID %d: %v", id, err)
		return fmt.Errorf("failed to process payment")
	}

	newAmount := debt.Amount - amount

	switch {
	case newAmount < 0:
		msg := fmt.Sprintf("payment amount exceeds debt. Maximum payment: %d", debt.Amount)
		s.logger.Warnf(msg+", Debt ID: %d", id)
		return fmt.Errorf(msg)

	case newAmount > 0:
		debt.Amount = newAmount
		if err = s.debtManager.UpdateDebt(ctx, debt); err != nil {
			s.logger.Errorf("failed to update debt after payment, ID %d: %v", id, err)
			return fmt.Errorf("failed to process payment")
		}
		s.logger.Infof("partial payment successful, Debt ID: %d, Remaining: %d", id, newAmount)
		return nil

	case newAmount == 0:
		if err = s.debtManager.DeleteDebt(ctx, id); err != nil {
			s.logger.Errorf("failed to delete fully paid debt, ID %d: %v", id, err)
			return fmt.Errorf("failed to complete payment")
		}
		s.logger.Infof("debt fully paid and deleted, ID: %d", id)
		return nil
	}

	return nil
}
