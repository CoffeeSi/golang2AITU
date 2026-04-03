package http

import (
	"errors"
	"net/http"

	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/transport/http/dto"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	uc usecase.AppointmentUsecase
}

func RegisterAppointmentHandlers(r *gin.Engine, uc usecase.AppointmentUsecase) {
	h := &AppointmentHandler{uc: uc}
	r.POST("/appointments", h.CreateAppointmentHandler)
	r.GET("/appointments/:id", h.GetAppointmentByIDHandler)
	r.GET("/appointments", h.ListAppointmentsHandler)
	r.PATCH("/appointments/:id/status", h.UpdateAppointmentStatusHandler)
}

func (h *AppointmentHandler) CreateAppointmentHandler(c *gin.Context) {
	var request dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if err := h.uc.CreateAppointment(request); err != nil {
		if errors.Is(err, usecase.ServiceUnavailableError) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "doctor service is temporarily unavailable"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "appointment created successfully"})
}

func (h *AppointmentHandler) GetAppointmentByIDHandler(c *gin.Context) {
	id := c.Param("id")

	appointment, err := h.uc.GetAppointmentByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if appointment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		return
	}
	c.JSON(http.StatusOK, appointment)
}

func (h *AppointmentHandler) ListAppointmentsHandler(c *gin.Context) {
	appointments, err := h.uc.ListAppointments()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(appointments) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no appointments found"})
		return
	}

	c.JSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) UpdateAppointmentStatusHandler(c *gin.Context) {
	id := c.Param("id")
	var request dto.UpdateAppointmentStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if err := h.uc.UpdateAppointmentStatus(id, request); err != nil {
		if errors.Is(err, usecase.ServiceUnavailableError) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "doctor service is temporarily unavailable"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "appointment status updated successfully"})
}
