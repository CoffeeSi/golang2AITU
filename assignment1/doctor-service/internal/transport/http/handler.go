package http

import (
	"net/http"

	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/transport/http/dto"
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type DoctorHandler struct {
	uc usecase.DoctorUsecase
}

func RegisterDoctorHandlers(r *gin.Engine, uc usecase.DoctorUsecase) {
	h := &DoctorHandler{uc: uc}
	r.POST("/doctors", h.CreateDoctorHandler)
	r.GET("/doctors/:id", h.GetDoctorByIDHandler)
	r.GET("/doctors", h.ListDoctorsHandler)
}

func (h *DoctorHandler) CreateDoctorHandler(c *gin.Context) {
	var request dto.CreateDoctorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if err := h.uc.CreateDoctor(request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "doctor created successfully"})
}

func (h *DoctorHandler) GetDoctorByIDHandler(c *gin.Context) {
	id := c.Param("id")

	doctor, err := h.uc.GetDoctorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if doctor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}
	c.JSON(http.StatusOK, doctor)
}

func (h *DoctorHandler) ListDoctorsHandler(c *gin.Context) {
	doctors, err := h.uc.ListDoctors()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} 
	if len(doctors) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no doctors found"})
		return
	}

	c.JSON(http.StatusOK, doctors)
}