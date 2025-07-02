package add

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
)

type TrendDeleter interface {
	DeleteTrend(ctx context.Context, name string) (int64, error)
}

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
			logger.Errorf("failed to save challenge: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to save challenge", logger)
			return
		}

		logger.Infof("trend (ID:%d) deleted successfully", id)
		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
