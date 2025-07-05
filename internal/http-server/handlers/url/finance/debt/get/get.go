package get

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
)

type DebtGetter interface {
	GetAmountDebt(ctx context.Context) (*model.AmountDebt, error)
}

// New
// @Summary Get aggregated debt information
// @Description Returns comprehensive debt information including:
// @Description - Total debt amount
// @Description - Detailed list of all debts with:
// @Description   - Debt ID
// @Description   - Description
// @Description  - Individual amount
// @Tags debts
// @Produce json
// @Success 200 {object} model.AmountDebt "Successful response with debt details"
// @Failure 404 {object} response.Response "No debts found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/finance/debt/get [get]
func New(ctx context.Context, getter DebtGetter, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := getter.GetAmountDebt(ctx)
		if err != nil {
			logger.Errorf("failed to get trend data: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to get trend data", logger)
			return
		}

		response.RespondWithJSON(w, http.StatusOK, result, logger)
	}
}
