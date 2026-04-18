package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type ProfileUseCaser interface {
	Get(ctx context.Context) (domain.HealthProfile, error)
	Upsert(ctx context.Context, profile domain.HealthProfile) (domain.HealthProfile, error)
}

type ProfileHandler struct {
	uc ProfileUseCaser
}

func NewProfileHandler(uc ProfileUseCaser) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

func (h *ProfileHandler) Register(g *echo.Group) {
	g.GET("/health/profile", h.GetProfile)
	g.PUT("/health/profile", h.UpsertProfile)
}

type profileRequest struct {
	HeightCM  float64 `json:"height_cm"`
	BirthDate string  `json:"birth_date"`
	Sex       string  `json:"sex"`
}

type profileResponse struct {
	ID        int     `json:"id"`
	HeightCM  float64 `json:"height_cm"`
	BirthDate string  `json:"birth_date"`
	Sex       string  `json:"sex"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (h *ProfileHandler) GetProfile(c echo.Context) error {
	profile, err := h.uc.Get(c.Request().Context())
	if err != nil {
		return mapError(c, err)
	}
	return c.JSON(http.StatusOK, toProfileResponse(profile))
}

func (h *ProfileHandler) UpsertProfile(c echo.Context) error {
	var req profileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse("invalid_request_body", "Request body is invalid JSON", nil))
	}
	if req.HeightCM <= 0 {
		return c.JSON(http.StatusBadRequest, newErrorResponse(
			"validation_failed",
			"Request validation failed",
			[]fieldError{{Field: "height_cm", Issue: "must be greater than 0"}},
		))
	}
	switch req.Sex {
	case "male", "female", "other":
	default:
		return c.JSON(http.StatusBadRequest, newErrorResponse(
			"validation_failed",
			"Request validation failed",
			[]fieldError{{Field: "sex", Issue: "must be one of: male, female, other"}},
		))
	}
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, newErrorResponse(
			"validation_failed",
			"Request validation failed",
			[]fieldError{{Field: "birth_date", Issue: "must be YYYY-MM-DD format"}},
		))
	}
	profile, err := h.uc.Upsert(c.Request().Context(), domain.HealthProfile{
		HeightCM:  req.HeightCM,
		BirthDate: birthDate.Format("2006-01-02"),
		Sex:       req.Sex,
	})
	if err != nil {
		return mapError(c, err)
	}
	return c.JSON(http.StatusOK, toProfileResponse(profile))
}

func toProfileResponse(profile domain.HealthProfile) profileResponse {
	return profileResponse{
		ID:        profile.ID,
		HeightCM:  profile.HeightCM,
		BirthDate: profile.BirthDate,
		Sex:       profile.Sex,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}
}
