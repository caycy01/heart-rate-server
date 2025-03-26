package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"heart-rate-server/internal/config"
	"heart-rate-server/internal/middleware"
	"heart-rate-server/internal/models"
	"heart-rate-server/internal/utils"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type App struct {
	DB           *gorm.DB
	Redis        *redis.Client
	Config       *config.Config
	SecureCookie *middleware.SecureCookie
}

var validate = validator.New()

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, err.Error(), "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, err, "Validation failed")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), app.Config.BcryptCost)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to hash password")
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		UUID:     uuid.New().String(),
	}

	result := app.DB.Create(&user)
	if result.Error != nil {
		utils.SendError(w, http.StatusConflict, result.Error, "Username already exists")
		return
	}

	//utils.SendResponse(w, http.StatusCreated, "User registered successfully", map[string]interface{}{
	//	"user_id": user.ID,
	//})

	authInfo := models.AuthInfo{
		UserID:   user.ID,
		Username: user.Username,
		Expires:  time.Now().Add(app.Config.TokenExpiry),
	}

	if err := app.SecureCookie.SetAuthCookie(w, authInfo); err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	utils.SendResponse(w, http.StatusOK, "registered successfully", nil)
}

func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, err, "Validation failed")
		return
	}

	var user models.User
	result := app.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.SendError(w, http.StatusUnauthorized, nil, "Invalid username or password")
		} else {
			utils.SendError(w, http.StatusInternalServerError, result.Error, "Database error")
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.SendError(w, http.StatusUnauthorized, nil, "Invalid username or password")
		return
	}

	authInfo := models.AuthInfo{
		UserID:   user.ID,
		Username: user.Username,
		Expires:  time.Now().Add(app.Config.TokenExpiry),
	}

	if err := app.SecureCookie.SetAuthCookie(w, authInfo); err != nil {
		utils.SendError(w, http.StatusInternalServerError, err, "Failed to create session")
		return
	}

	utils.SendResponse(w, http.StatusOK, "Logged in successfully", nil)
}

func (app *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	app.SecureCookie.ClearAuthCookie(w)
	utils.SendResponse(w, http.StatusOK, "Logged out successfully", nil)
}
