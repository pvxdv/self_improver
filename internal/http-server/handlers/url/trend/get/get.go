package get

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
	"github.com/pvxdv/self_improver/internal/storage"
)

type TrendGetter interface {
	GetTrend(ctx context.Context, name string) (*model.Trend, error)
}

// New
// @Summary Get trend with challenges
// @Description Returns a trend and all associated challenges
// @Tags trends
// @Produce json
// @Param name query string true "Trend name"
// @Success 200 {object} model.Trend
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/trend/get [get]
func New(ctx context.Context, getter TrendGetter, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trendName := r.URL.Query().Get("name")
		if trendName == "" {
			logger.Warnf("trend name not provided")
			response.RespondWithError(w, http.StatusBadRequest, "trend name parameter is required", logger)
			return
		}

		trend, err := getter.GetTrend(ctx, trendName)
		if err != nil {
			if errors.Is(err, storage.ErrTrendNotFound) {
				logger.Warnf("trend not found: %s", trendName)
				response.RespondWithError(w, http.StatusNotFound, "trend not found", logger)
				return
			}

			logger.Errorf("failed to get trend data: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to get trend data", logger)
			return
		}

		logger.Infof("successfully retrieved trend %s", trendName)
		response.RespondWithJSON(w, http.StatusOK, trend, logger)
	}
}
