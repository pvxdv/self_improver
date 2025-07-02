package delete

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/storage"
)

type TrendDeleter interface {
	DeleteTrend(ctx context.Context, name string) (int64, error)
}

// New
// @Summary Delete trend by name
// @Description Removes trend and all associated challenges
// @Tags trends
// @Produce json
// @Param name query string true "Trend name"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/trend/delete [delete]
func New(ctx context.Context, deleter TrendDeleter, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trendName := r.URL.Query().Get("name")
		if trendName == "" {
			logger.Warnf("trend name not provided")
			response.RespondWithError(w, http.StatusBadRequest, "trend name parameter is required", logger)
			return
		}

		id, err := deleter.DeleteTrend(ctx, trendName)
		if err != nil {
			logger.Errorf("failed to delete trend: %v", err)

			if errors.Is(storage.ErrTrendNotFound, err) {
				response.RespondWithError(w, http.StatusInternalServerError, err.Error(), logger)
				return
			}

			response.RespondWithError(w, http.StatusInternalServerError, "failed to delete trend", logger)
			return
		}

		logger.Infof("trend (ID:%d) deleted successfully", id)
		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
