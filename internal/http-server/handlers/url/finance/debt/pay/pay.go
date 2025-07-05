package delete

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
	"github.com/pvxdv/self_improver/internal/storage"
)

type DebtPayer interface {
	GetDebt(ctx context.Context, id int) (model.Debt, error)
	UpdateDebt(ctx context.Context, debt model.Debt) error
	DeleteDebt(ctx context.Context, id int64) error
}

// New
// @Summary Process debt payment
// @Description Handles debt payment with three possible outcomes:
// @Description  - Partial payment: reduces debt amount
// @Description  - Full payment: deletes the debt
// @Description  - Overpayment: rejects with 409 Conflict
// @Tags debts
// @Accept json
// @Produce json
// @Param id query int true "Debt ID" example(123)
// @Param amount query int true "Payment amount" example(5000)
// @Success 200 {object} response.Response "Debt updated or deleted successfully"
// @Failure 400 {object} response.Response "Invalid/missing parameters"
// @Failure 404 {object} response.Response "Debt not found"
// @Failure 409 {object} response.Response "Payment amount exceeds debt"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/finance/debt/pay [post]
func New(ctx context.Context, payer DebtPayer, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			logger.Warn("id parameter is required")
			response.RespondWithError(w, http.StatusBadRequest, "id parameter is required", logger)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Warnf("invalid id parameter: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "id parameter must be a valid integer", logger)
			return
		}

		amountStr := r.URL.Query().Get("amount")
		if amountStr == "" {
			logger.Warn("amount parameter is required")
			response.RespondWithError(w, http.StatusBadRequest, "amount parameter is required", logger)
			return
		}

		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			logger.Warnf("invalid amount parameter: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "amount parameter must be a valid integer", logger)
			return
		}

		debt, err := payer.GetDebt(ctx, id)
		if err != nil {
			if errors.Is(storage.ErrDebtNotFound, err) {
				logger.Warnf("debt not found: %v", err)
				response.RespondWithError(w, http.StatusNotFound, "debt not found", logger)
				return
			}

			logger.Errorf("failed to get debt: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to retrieve debt", logger)
			return
		}

		newAmount := debt.Amount - int64(amount)

		switch {
		case newAmount < 0:
			msg := fmt.Sprintf("payment amount exceeds debt. Maximum payment: %d", debt.Amount)
			logger.Warn(msg)
			response.RespondWithError(w, http.StatusConflict, msg, logger)
			return

		case newAmount > 0:
			debt.Amount = newAmount
			if err = payer.UpdateDebt(ctx, debt); err != nil {
				logger.Errorf("failed to update debt: %v", err)
				response.RespondWithError(w, http.StatusInternalServerError, "failed to update debt", logger)
				return
			}
			logger.Infof("debt partially paid. ID: %d, Remaining amount: %d", id, newAmount)
			response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)

		case newAmount == 0:
			if err = payer.DeleteDebt(ctx, int64(id)); err != nil {
				logger.Errorf("failed to delete debt: %v", err)
				response.RespondWithError(w, http.StatusInternalServerError, "failed to delete debt", logger)
				return
			}
			logger.Infof("debt fully paid and deleted. ID: %d", id)
			response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
		}
	}
}
