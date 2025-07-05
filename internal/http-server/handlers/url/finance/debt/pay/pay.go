package delete

import (
	"context"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
)

type DebtPayer interface {
	PayDebt(ctx context.Context, id int64, amount int64) error
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

		err = payer.PayDebt(ctx, int64(id), int64(amount))
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "failed to retrieve debt", logger)
			return
		}

		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
