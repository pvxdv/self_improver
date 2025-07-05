package update

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
)

type DebtUpdater interface {
	UpdateDebt(ctx context.Context, debt model.Debt) error
}

// New
// @Summary Update existing debt
// @Description Updates an existing debt record by ID. Requires full debt data including:
// @Description  - ID (mandatory)
// @Description  - New amount
// @Description  - Description
// @Tags debts
// @Accept json
// @Produce json
// @Param input body model.Debt true "Debt update data" example({"id": 1, "amount": 1500000, "description": "Updated car loan amount"})
// @Success 200 {object} response.Response "Debt updated successfully"
// @Failure 400 {object} response.Response "Invalid request format or missing ID"
// @Failure 404 {object} response.Response "Debt not found for given ID"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/finance/debt/update [put]
func New(ctx context.Context, updater DebtUpdater, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("failed to read request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "failed to read request body", logger)
			return
		}

		var debt model.Debt
		if err := json.Unmarshal(data, &debt); err != nil {
			logger.Errorf("failed to unmarshal request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "invalid request format", logger)
			return
		}

		logger.Debugf("resived debt to update:%v", debt)

		err = updater.UpdateDebt(ctx, debt)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error(), logger)
			return
		}

		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
