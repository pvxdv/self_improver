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

type ChallengeSaver interface {
	SaveChallenge(ctx context.Context, data *model.Challenge) (int64, error)
}

// New
// @Summary Add new challenge
// @Description Creates a new challenge under an existing trend
// @Tags challenges
// @Accept json
// @Produce json
// @Param input body model.Challenge true "Challenge data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/challenge/add [post]
func New(ctx context.Context, saver ChallengeSaver, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("failed to read request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "failed to read request body", logger)
			return
		}

		var challenge model.Challenge
		if err := json.Unmarshal(data, &challenge); err != nil {
			logger.Errorf("failed to unmarshal request body: %v", err)
			response.RespondWithError(w, http.StatusBadRequest, "invalid request format", logger)
			return
		}

		//TODO validate

		id, err := saver.SaveChallenge(ctx, &challenge)
		if err != nil {
			logger.Errorf("failed to save challenge: %v", err)

			if errors.Is(storage.ErrTrendNotFound, err) {
				response.RespondWithError(w, http.StatusInternalServerError, err.Error(), logger)
				return
			}

			response.RespondWithError(w, http.StatusInternalServerError, "failed to save challenge", logger)
			return
		}

		logger.Infof("challenge (ID:%d) saved successfully", id)
		response.RespondWithJSON(w, http.StatusOK, response.OK(), logger)
	}
}
