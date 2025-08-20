package api

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/api/repository"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/update"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	"github.com/daniil11ru/egts/cli/receiver/dto/other"
	"github.com/daniil11ru/egts/cli/receiver/dto/request"
	"github.com/daniil11ru/egts/cli/receiver/dto/response"
	"github.com/daniil11ru/egts/cli/receiver/util"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type Handler struct {
	Repository repository.BusinessData
}

func NewHandler(repository repository.BusinessData) *Handler {
	return &Handler{Repository: repository}
}

func (h *Handler) GetVehicles(c *gin.Context) {
	getVehiclesFilter := filter.Vehicles{}

	if providerIdStr := c.Query("provider_id"); providerIdStr != "" {
		providerId, err := strconv.Atoi(providerIdStr)
		if err == nil {
			if providerId < 0 || providerId > 2147483647 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "provider_id должен быть в пределах от 0 до 2147483647"})
				return
			}
			providerId32 := int32(providerId)
			getVehiclesFilter.ProviderId = &providerId32
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный provider_id"})
			return
		}
	}

	if moderationStatusStr := c.Query("moderation_status"); moderationStatusStr != "" {
		moderationStatus := other.ModerationStatus(moderationStatusStr)
		if !moderationStatus.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный moderation_status"})
			return
		}
		getVehiclesFilter.ModerationStatus = &moderationStatus
	}

	if imei := c.Query("imei"); imei != "" {
		getVehiclesFilter.IMEI = &imei
	}

	vehicles, err := h.Repository.GetVehicles(getVehiclesFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := response.GetVehicles(util.Map(vehicles, func(item out.Vehicle) response.GetVehicle {
		return response.GetVehicle{
			ID:               item.ID,
			IMEI:             item.IMEI,
			OID:              item.OID,
			Name:             item.Name,
			ProviderID:       item.ProviderId,
			ModerationStatus: item.ModerationStatus.String(),
		}
	}))

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetVehicle(c *gin.Context) {
	vehicleIdStr := c.Param("id")
	vehicleId, err := strconv.ParseInt(vehicleIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID транспорта"})
		return
	}

	vehicle, err := h.Repository.GetVehicle(int32(vehicleId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	resp := response.GetVehicle{
		ID:               vehicle.ID,
		IMEI:             vehicle.IMEI,
		OID:              vehicle.OID,
		Name:             vehicle.Name,
		ProviderID:       vehicle.ProviderId,
		ModerationStatus: vehicle.ModerationStatus.String(),
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetVehiclesExcel(c *gin.Context) {
	getVehiclesFilter := filter.Vehicles{}

	if s := c.Query("provider_id"); s != "" {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некоррректный provider_id"})
			return
		}
		if v < 0 || v > 2147483647 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider_id должен быть в пределах от 0 до 2147483647"})
			return
		}
		id := int32(v)
		getVehiclesFilter.ProviderId = &id
	}

	if s := c.Query("moderation_status"); s != "" {
		ms := other.ModerationStatus(s)
		if !ms.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный moderation_status"})
			return
		}
		getVehiclesFilter.ModerationStatus = &ms
	}

	vehicles, err := h.Repository.GetVehicles(getVehiclesFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	headers := []string{"ID", "IMEI", "OID", "Название", "ID провайдера", "Статус модерации"}
	for i, hName := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, hName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	for i, v := range vehicles {
		row := i + 2
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), v.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("B%d", row), v.IMEI); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if v.OID != nil {
			if err := f.SetCellValue(sheet, fmt.Sprintf("C%d", row), *v.OID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if v.Name != nil {
			if err := f.SetCellValue(sheet, fmt.Sprintf("D%d", row), *v.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("E%d", row), v.ProviderId); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("F%d", row), v.ModerationStatus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=vehicles.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *Handler) GetLocations(c *gin.Context) {
	getLocationsFilter := filter.Locations{}

	const timeLayout = "02.01.2006 15:04:05"

	if locationsLimitStr := c.Query("locations_limit"); locationsLimitStr != "" {
		locationsLimit, err := strconv.Atoi(locationsLimitStr)
		if err == nil {
			if locationsLimit < 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "locations_limit должен быть в пределах от 1 до 9223372036854775807"})
				return
			}
			getLocationsFilter.LocationsLimit = int64(locationsLimit)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный locations_limit"})
			return
		}
	} else {
		getLocationsFilter.LocationsLimit = 10
	}

	if vehicleIdStr := c.Query("vehicle_id"); vehicleIdStr != "" {
		vehicleId, err := strconv.Atoi(vehicleIdStr)
		if err == nil {
			if vehicleId < 0 || vehicleId > 2147483647 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id должен быть в пределах от 0 до 2147483647"})
				return
			}
			vehicleId32 := int32(vehicleId)
			getLocationsFilter.VehicleId = &vehicleId32
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный vehicle_id"})
			return
		}
	}
	if sentBeforeStr := c.Query("sent_before"); sentBeforeStr != "" {
		sentBefore, err := time.Parse(timeLayout, sentBeforeStr)
		if err == nil {
			getLocationsFilter.SentBefore = &sentBefore
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Дата должна быть в формате DD.MM.YYYY HH:MM:SS"})
			return
		}
	}
	if sentAfterStr := c.Query("sent_after"); sentAfterStr != "" {
		sentAfter, err := time.Parse(timeLayout, sentAfterStr)
		if err == nil {
			getLocationsFilter.SentAfter = &sentAfter
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Дата должна быть в формате DD.MM.YYYY HH:MM:SS"})
			return
		}
	}
	if receivedBeforeStr := c.Query("received_before"); receivedBeforeStr != "" {
		receivedBefore, err := time.Parse(timeLayout, receivedBeforeStr)
		if err == nil {
			getLocationsFilter.ReceivedBefore = &receivedBefore
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Дата должна быть в формате DD.MM.YYYY HH:MM:SS"})
			return
		}
	}
	if receivedAfterStr := c.Query("received_after"); receivedAfterStr != "" {
		receivedAfter, err := time.Parse(timeLayout, receivedAfterStr)
		if err == nil {
			getLocationsFilter.ReceivedAfter = &receivedAfter
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Дата должна быть в формате DD.MM.YYYY HH:MM:SS"})
			return
		}
	}

	locations, err := h.Repository.GetLocations(getLocationsFilter)
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
		sentAt := loc.SentAt.Format(timeLayout)
		track.Locations = append(track.Locations, response.Location{
			OID:            loc.OID,
			Latitude:       loc.Latitude,
			Longitude:      loc.Longitude,
			Altitude:       loc.Altitude,
			Direction:      loc.Direction,
			Speed:          loc.Speed,
			SatelliteCount: loc.SatelliteCount,
			SentAt:         &sentAt,
			ReceivedAt:     loc.ReceivedAt.Format(timeLayout)})
		tracks[loc.VehicleId] = track
	}

	vehicleTracks := make([]response.VehicleTrack, 0, len(tracks))
	for _, track := range tracks {
		vehicleTracks = append(vehicleTracks, track)
	}

	c.JSON(http.StatusOK, response.GetLocations(vehicleTracks))
}

func (h *Handler) UpdateVehicleByImei(c *gin.Context) {
	var req request.UpdateVehicle
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IMEI == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IMEI обязателен для указания"})
		return
	}
	if req.Name == nil && req.ModerationStatus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нечего обновлять"})
		return
	}
	if err := h.Repository.UpdateVehicleByImei(*req.IMEI, update.VehicleByImei{
		Name:             req.Name,
		ModerationStatus: req.ModerationStatus,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) UpdateVehicleById(c *gin.Context) {
	var req request.UpdateVehicle

	vehicleIdStr := c.Param("id")
	vehicleId, err := strconv.ParseInt(vehicleIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID транспорта"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IMEI == nil && req.Name == nil && req.ModerationStatus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нечего обновлять"})
		return
	}
	if err := h.Repository.UpdateVehicleById(int32(vehicleId), update.VehicleById{
		Name:             req.Name,
		ModerationStatus: req.ModerationStatus,
		IMEI:             req.IMEI,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
