package add

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/pvxdv/self_improver/internal/lib/api/response"
	"github.com/pvxdv/self_improver/internal/model"
	"github.com/pvxdv/self_improver/internal/storage"
)

type TrendSaver interface {
	SaveTrend(ctx context.Context, data *model.Trend) (int64, error)
}

// New
// @Summary Create new trend
// @Description Adds a new trend with name
// @Tags trends
// @Accept json
// @Produce json
// @Param input body model.Trend true "Trend data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/trend/add [post]
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

			if errors.Is(storage.ErrTrendExists, err) {
				response.RespondWithError(w, http.StatusInternalServerError, err.Error(), logger)
				return
			}

			response.RespondWithError(w, http.StatusInternalServerError, "failed to save trend", logger)
			return
		}

		logger.Infof("trend (ID:%d) saved successfully", id)
		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
