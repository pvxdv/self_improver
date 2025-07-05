package add

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
)

type DeptAdder interface {
	AddDebt(ctx context.Context, data *model.Debt) error
}

// New
// @Summary Create new debt
// @Description Creates a new debt record in the system.
// @Description Accepts JSON with debt details
// @Tags debts
// @Accept json
// @Produce json
// @Param input body model.Debt true "Debt creation data"
// @Success 200 {object} response.Response "Returns success message with debt ID"
// @Failure 400 {object} response.Response "Invalid request format or missing required fields"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/finance/debt/add [post]
func New(ctx context.Context, adder DeptAdder, logger *zap.SugaredLogger) http.HandlerFunc {
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

		logger.Debugf("resived debt to save:%v", debt)

		err = adder.AddDebt(ctx, &debt)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error(), logger)
			return
		}

		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
