package api

import (
	"net/http"
	"strconv"

	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
	"github.com/gin-gonic/gin"
)

type Repository interface {
	GetVehicles(providerId int32) ([]types.Vehicle, error)
}

type Handler struct {
	Repository Repository
}

func New(repository Repository) *Handler {
	return &Handler{Repository: repository}
}

func (h *Handler) GetVehicles(c *gin.Context) {
	providerId, _ := strconv.Atoi(c.Param("provider_id"))
	vehicles, err := h.Repository.GetVehicles(int32(providerId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vehicles)
}
