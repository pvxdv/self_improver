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

type TrendSaver interface {
	SaveTrend(ctx context.Context, data *model.Trend) (int64, error)
}

func New(ctx context.Context, saver TrendSaver, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("failed to read request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "failed to read request body", logger)
			return
		}

		var challenge model.Trend
		if err := json.Unmarshal(data, &challenge); err != nil {
			logger.Errorf("failed to unmarshal request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "invalid request format", logger)
			return
		}

		//TODO validate

		id, err := saver.SaveTrend(ctx, &challenge)
		if err != nil {
			logger.Errorf("failed to save trend: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to save trend", logger)
			return
		}

		logger.Infof("trend (ID:%d) saved successfully", id)
		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
