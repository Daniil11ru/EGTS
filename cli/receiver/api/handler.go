package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/api/dto/request"
	"github.com/daniil11ru/egts/cli/receiver/api/dto/response"
	"github.com/daniil11ru/egts/cli/receiver/api/model"
	"github.com/daniil11ru/egts/cli/receiver/api/repository"
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
	"github.com/daniil11ru/egts/cli/receiver/util"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository repository.BusinessData
}

func NewHandler(repository repository.BusinessData) *Handler {
	return &Handler{Repository: repository}
}

func (h *Handler) GetVehicles(c *gin.Context) {
	request := request.GetVehicles{}

	if providerIdStr := c.Query("provider_id"); providerIdStr != "" {
		providerId, err := strconv.Atoi(providerIdStr)
		if err == nil {
			providerId32 := int32(providerId)
			request.ProviderID = &providerId32
		}
	}

	if moderationStatusStr := c.Query("moderation_status"); moderationStatusStr != "" {
		moderationStatus := types.ModerationStatus(moderationStatusStr)
		request.ModerationStatus = &moderationStatus
	}

	if imeiStr := c.Query("imei"); imeiStr != "" {
		imei, err := strconv.ParseInt(imeiStr, 10, 64)
		if err == nil {
			request.IMEI = &imei
		}
	}

	vehicles, err := h.Repository.GetVehicles(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := util.Map(vehicles, func(item model.Vehicle) response.Vehicle {
		return response.Vehicle{
			ID:               item.ID,
			IMEI:             strconv.FormatInt(item.IMEI, 10),
			OID:              item.OID,
			Name:             item.Name,
			ProviderID:       item.ProviderID,
			ModerationStatus: item.ModerationStatus,
		}
	})

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetLocations(c *gin.Context) {
	request := request.GetLocations{}

	const timeLayout = "02.01.2006 15:04:05"

	if vehicleIdStr := c.Query("vehicle_id"); vehicleIdStr != "" {
		vehicleId, err := strconv.Atoi(vehicleIdStr)
		if err == nil {
			vehicleId32 := int32(vehicleId)
			request.VehicleID = &vehicleId32
		}
	}
	if sentBeforeStr := c.Query("sent_before"); sentBeforeStr != "" {
		sentBefore, err := time.Parse(timeLayout, sentBeforeStr)
		if err == nil {
			request.SentBefore = &sentBefore
		}
	}
	if sentAfterStr := c.Query("sent_after"); sentAfterStr != "" {
		sentAfter, err := time.Parse(timeLayout, sentAfterStr)
		if err == nil {
			request.SentAfter = &sentAfter
		}
	}
	if receivedBeforeStr := c.Query("received_before"); receivedBeforeStr != "" {
		receivedBefore, err := time.Parse(timeLayout, receivedBeforeStr)
		if err == nil {
			request.ReceivedBefore = &receivedBefore
		}
	}
	if receivedAfterStr := c.Query("received_after"); receivedAfterStr != "" {
		receivedAfter, err := time.Parse(timeLayout, receivedAfterStr)
		if err == nil {
			request.ReceivedAfter = &receivedAfter
		}
	}

	locations, err := h.Repository.GetLocations(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tracks := make(map[int32]response.VehicleTrack)
	for _, loc := range locations {
		track, exists := tracks[loc.VehicleId]
		if !exists {
			track = response.VehicleTrack{
				VehicleId: loc.VehicleId,
				Locations: []response.Location{},
			}
		}
		track.Locations = append(track.Locations, response.Location{Latitude: loc.Latitude, Longitude: loc.Longitude, SentAt: loc.SentAt, ReceivedAt: loc.ReceivedAt})
		tracks[loc.VehicleId] = track
	}

	vehicleTracks := make([]response.VehicleTrack, 0, len(tracks))
	for _, track := range tracks {
		vehicleTracks = append(vehicleTracks, track)
	}

	c.JSON(http.StatusOK, vehicleTracks)
}

func (h *Handler) GetLatestLocations(c *gin.Context) {
	request := request.GetLatestLocations{}

	if vehicleIdStr := c.Query("vehicle_id"); vehicleIdStr != "" {
		vehicleId, err := strconv.Atoi(vehicleIdStr)
		if err == nil {
			vehicleId32 := int32(vehicleId)
			request.VehicleID = &vehicleId32
		}
	}

	locations, err := h.Repository.GetLatestLocations(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(locations) == 0 {
		locations = make([]response.LatestLocation, 0)
	}

	c.JSON(http.StatusOK, locations)
}

func (h *Handler) UpdateVehicle(c *gin.Context) {
	var req request.UpdateVehicle
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == nil && req.ModerationStatus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	if err := h.Repository.UpdateVehicle(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
