package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"github.com/rafaelsoares/alfredo/internal/logger"
	"go.uber.org/zap"
)

type PetServicer interface {
	List(ctx context.Context) ([]domain.Pet, error)
	Create(ctx context.Context, in service.CreatePetInput) (*domain.Pet, error)
	GetByID(ctx context.Context, id string) (*domain.Pet, error)
	Update(ctx context.Context, id string, in service.UpdatePetInput) (*domain.Pet, error)
	Delete(ctx context.Context, id string) error
}

type PetHandler struct {
	svc PetServicer
}

func NewPetHandler(svc PetServicer) *PetHandler {
	return &PetHandler{svc: svc}
}

func (h *PetHandler) Register(g *echo.Group) {
	g.GET("/pets", h.List)
	g.POST("/pets", h.Create)
	g.GET("/pets/:id", h.GetByID)
	g.PUT("/pets/:id", h.Update)
	g.DELETE("/pets/:id", h.Delete)
}

type petRequest struct {
	Name           string   `json:"name"`
	Species        string   `json:"species"`
	Breed          *string  `json:"breed"`
	BirthDate      *string  `json:"birth_date"`
	WeightKg       *float64 `json:"weight_kg"`
	DailyFoodGrams *float64 `json:"daily_food_grams"`
	PhotoPath      *string  `json:"photo_path"`
}

type petResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Species        string   `json:"species"`
	Breed          *string  `json:"breed,omitempty"`
	BirthDate      *string  `json:"birth_date,omitempty"`
	WeightKg       *float64 `json:"weight_kg,omitempty"`
	DailyFoodGrams *float64 `json:"daily_food_grams,omitempty"`
	PhotoPath      *string  `json:"photo_path,omitempty"`
	CreatedAt      string   `json:"created_at"`
}

func toPetResponse(p domain.Pet) petResponse {
	r := petResponse{
		ID:             p.ID,
		Name:           p.Name,
		Species:        p.Species,
		Breed:          p.Breed,
		WeightKg:       p.WeightKg,
		DailyFoodGrams: p.DailyFoodGrams,
		PhotoPath:      p.PhotoPath,
		CreatedAt:      p.CreatedAt.Format(time.RFC3339),
	}
	if p.BirthDate != nil {
		s := p.BirthDate.Format("2006-01-02")
		r.BirthDate = &s
	}
	return r
}

func (h *PetHandler) List(c echo.Context) error {
	pets, err := h.svc.List(c.Request().Context())
	if err != nil {
		return mapError(c, err)
	}
	resp := make([]petResponse, 0, len(pets))
	for _, p := range pets {
		resp = append(resp, toPetResponse(p))
	}
	logger.FromEcho(c).Info("pets listed", zap.Int("count", len(pets)))
	return c.JSON(http.StatusOK, resp)
}

func (h *PetHandler) Create(c echo.Context) error {
	var req petRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorBody("invalid_request_body"))
	}
	var birthDate *time.Time
	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorBody("invalid birth_date format, use YYYY-MM-DD"))
		}
		birthDate = &t
	}
	pet, err := h.svc.Create(c.Request().Context(), service.CreatePetInput{
		Name: req.Name, Species: req.Species, Breed: req.Breed,
		BirthDate: birthDate, WeightKg: req.WeightKg,
		DailyFoodGrams: req.DailyFoodGrams, PhotoPath: req.PhotoPath,
	})
	if err != nil {
		return mapError(c, err)
	}
	logger.FromEcho(c).Info("pet created", zap.String("pet_id", pet.ID), zap.String("pet_name", pet.Name), zap.String("species", pet.Species))
	return c.JSON(http.StatusCreated, toPetResponse(*pet))
}

func (h *PetHandler) GetByID(c echo.Context) error {
	pet, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapError(c, err)
	}
	logger.FromEcho(c).Info("pet fetched", zap.String("pet_id", pet.ID), zap.String("pet_name", pet.Name))
	return c.JSON(http.StatusOK, toPetResponse(*pet))
}

func (h *PetHandler) Update(c echo.Context) error {
	var req petRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorBody("invalid_request_body"))
	}
	var birthDate *time.Time
	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorBody("invalid birth_date format, use YYYY-MM-DD"))
		}
		birthDate = &t
	}
	pet, err := h.svc.Update(c.Request().Context(), c.Param("id"), service.UpdatePetInput{
		Name: req.Name, Species: req.Species, Breed: req.Breed,
		BirthDate: birthDate, WeightKg: req.WeightKg,
		DailyFoodGrams: req.DailyFoodGrams, PhotoPath: req.PhotoPath,
	})
	if err != nil {
		return mapError(c, err)
	}
	logger.FromEcho(c).Info("pet updated", zap.String("pet_id", pet.ID), zap.String("pet_name", pet.Name))
	return c.JSON(http.StatusOK, toPetResponse(*pet))
}

func (h *PetHandler) Delete(c echo.Context) error {
	petID := c.Param("id")
	if err := h.svc.Delete(c.Request().Context(), petID); err != nil {
		return mapError(c, err)
	}
	logger.FromEcho(c).Info("pet deleted", zap.String("pet_id", petID))
	return c.NoContent(http.StatusNoContent)
}

