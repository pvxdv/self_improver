package delete

import (
	"context"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
)

type DebtDeleter interface {
	DeleteDebt(ctx context.Context, id int64) error
}

// New
// @Summary Delete debt by id
// @Description Removes debt by id
// @Tags debts
// @Produce json
// @Param id query int true "id"
// @Success 200 {object} response.Response "Returns success message"
// @Failure 400 {object} response.Response "Invalid/missing parameters"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/finance/debt/delete [delete]
func New(ctx context.Context, deleter DebtDeleter, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			logger.Warnf("debt id not provided")
			response.RespondWithError(w, http.StatusBadRequest, "debt id parameter is required", logger)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Warnf("failed to convert debt id")
			response.RespondWithError(w, http.StatusBadRequest, "debt id parameter is invalid", logger)
		}

		err = deleter.DeleteDebt(ctx, int64(id))
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "failed to delete debt", logger)
			return
		}

		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
