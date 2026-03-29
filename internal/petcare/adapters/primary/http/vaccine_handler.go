package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/logger"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"go.uber.org/zap"
)

// VaccineServicer is the consumer-defined interface consumed by VaccineHandler.
type VaccineServicer interface {
	ListVaccines(ctx context.Context, petID string) ([]domain.Vaccine, error)
	RecordVaccine(ctx context.Context, in service.RecordVaccineInput) (*domain.Vaccine, error)
	DeleteVaccine(ctx context.Context, petID, vaccineID string) error
}

type VaccineHandler struct {
	svc VaccineServicer
}

func NewVaccineHandler(svc VaccineServicer) *VaccineHandler {
	return &VaccineHandler{svc: svc}
}

func (h *VaccineHandler) Register(g *echo.Group) {
	g.GET("/pets/:id/vaccines", h.ListVaccines)
	g.POST("/pets/:id/vaccines", h.RecordVaccine)
	g.DELETE("/pets/:id/vaccines/:vid", h.DeleteVaccine)
}

func (h *VaccineHandler) ListVaccines(c echo.Context) error {
	vs, err := h.svc.ListVaccines(c.Request().Context(), c.Param("id"))
	if err != nil {
		return mapError(c, err)
	}
	resp := make([]map[string]any, 0, len(vs))
	for _, v := range vs {
		resp = append(resp, vaccineToMap(v))
	}
	logger.FromEcho(c).Info("vaccines listed", zap.String("pet_id", c.Param("id")), zap.Int("count", len(vs)))
	return c.JSON(http.StatusOK, resp)
}

func (h *VaccineHandler) RecordVaccine(c echo.Context) error {
	var req struct {
		Name        string  `json:"name"`
		Date        string  `json:"date"`
		NextDueAt   *string `json:"next_due_at"`
		VetName     *string `json:"vet_name"`
		BatchNumber *string `json:"batch_number"`
		Notes       *string `json:"notes"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorBody("invalid_request_body"))
	}
	adminAt, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorBody("invalid date, use RFC3339"))
	}
	var nextDue *time.Time
	if req.NextDueAt != nil {
		t, err := time.Parse("2006-01-02", *req.NextDueAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorBody("invalid next_due_at, use YYYY-MM-DD"))
		}
		nextDue = &t
	}
	v, err := h.svc.RecordVaccine(c.Request().Context(), service.RecordVaccineInput{
		PetID: c.Param("id"), Name: req.Name, AdministeredAt: adminAt,
		NextDueAt: nextDue, VetName: req.VetName, BatchNumber: req.BatchNumber, Notes: req.Notes,
	})
	if err != nil {
		return mapError(c, err)
	}
	logger.FromEcho(c).Info("vaccine recorded", zap.String("pet_id", v.PetID), zap.String("vaccine_id", v.ID), zap.String("name", v.Name))
	return c.JSON(http.StatusCreated, vaccineToMap(*v))
}

func (h *VaccineHandler) DeleteVaccine(c echo.Context) error {
	if err := h.svc.DeleteVaccine(c.Request().Context(), c.Param("id"), c.Param("vid")); err != nil {
		return mapError(c, err)
	}
	logger.FromEcho(c).Info("vaccine deleted", zap.String("pet_id", c.Param("id")), zap.String("vaccine_id", c.Param("vid")))
	return c.NoContent(http.StatusNoContent)
}

// --- response helper ---

func vaccineToMap(v domain.Vaccine) map[string]any {
	m := map[string]any{"id": v.ID, "pet_id": v.PetID, "name": v.Name, "date": v.AdministeredAt.Format(time.RFC3339), "vet_name": v.VetName, "batch_number": v.BatchNumber, "notes": v.Notes}
	if v.NextDueAt != nil {
		m["next_due_at"] = v.NextDueAt.Format("2006-01-02")
	}
	return m
}
