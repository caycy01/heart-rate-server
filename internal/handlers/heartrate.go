package handlers

import (
	"encoding/json"
	"fmt"
	"heart-rate-server/internal/models"
	"heart-rate-server/internal/utils"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

func (app *App) ReceiveDataHandler(w http.ResponseWriter, r *http.Request) {
	authInfo := r.Context().Value("authInfo").(*models.AuthInfo)
	userID := fmt.Sprintf("%d", authInfo.UserID)

	var data models.HeartRateData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if err := validate.Struct(data); err != nil {
		utils.SendError(w, http.StatusBadRequest, err, "Validation failed")
		return
	}

	measuredTime := time.Unix(0, data.MeasuredAt*int64(time.Millisecond))
	now := time.Now()

	// Validate timestamp
	if measuredTime.After(now.Add(5 * time.Minute)) {
		utils.SendError(w, http.StatusBadRequest, nil, "Measurement time cannot be in the future")
		return
	}

	if measuredTime.Before(now.Add(-10 * time.Minute)) {
		utils.SendError(w, http.StatusBadRequest, nil, "Measurement time is too old")
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to process data")
		return
	}

	ctx := r.Context()
	key := fmt.Sprintf("heart_rate:%s", userID)
	z := &redis.Z{
		Score:  float64(data.MeasuredAt),
		Member: jsonData,
	}

	// Use pipeline for atomic operations
	pipe := app.Redis.TxPipeline()
	pipe.ZAdd(ctx, key, z)
	pipe.Expire(ctx, key, 10*time.Minute)
	_, err = pipe.Exec(ctx)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to store data")
		return
	}

	utils.SendResponse(w, http.StatusOK, "OK", nil)
}

func (app *App) LatestHeartRateHandler(w http.ResponseWriter, r *http.Request) {
	authInfo := r.Context().Value("authInfo").(*models.AuthInfo)
	userID := fmt.Sprintf("%d", authInfo.UserID)
	ctx := r.Context()
	key := fmt.Sprintf("heart_rate:%s", userID)

	// Check if key exists
	exists, err := app.Redis.Exists(ctx, key).Result()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to check data existence")
		return
	}
	if exists == 0 {
		utils.SendError(w, http.StatusNotFound, nil, "No heart rate data found")
		return
	}

	// Get latest data
	result, err := app.Redis.ZRevRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to retrieve data")
		return
	}

	if len(result) == 0 {
		utils.SendError(w, http.StatusNotFound, nil, "No heart rate data found")
		return
	}

	var storedData models.HeartRateData
	if err := json.Unmarshal([]byte(result[0].Member.(string)), &storedData); err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to parse data")
		return
	}

	utils.SendResponse(w, http.StatusOK, "", storedData)
}
