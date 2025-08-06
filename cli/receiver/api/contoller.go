package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daniil11ru/egts/cli/receiver/api/domain"
	"github.com/daniil11ru/egts/cli/receiver/api/model"
	"github.com/daniil11ru/egts/cli/receiver/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	Handler      *Handler
	Router       *gin.Engine
	GetApiKeys   *domain.GetApiKeys
	ApiKeyHashes []string
}

func NewController(handler *Handler, getApiKeys *domain.GetApiKeys) (*Controller, error) {
	ApiKeyAttributes, err := getApiKeys.Run()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации об API-ключах из базы данных: %w", err)
	}
	apiKeyHashes := util.Map(ApiKeyAttributes, func(item model.ApiKey) string {
		return item.Hash
	})

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		AllowCredentials: false,
	}))

	router.Use(func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		for _, apiKeyHashString := range apiKeyHashes {
			apiKeyHash, err := hex.DecodeString(apiKeyHashString)
			if err != nil {
				logrus.Errorf("ошибка декодирования хеш-строки API-ключа: %v", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			receivedApiKeyHash := sha256.Sum256([]byte(apiKey))
			computedReceivedApiKeyHash := receivedApiKeyHash[:]

			if subtle.ConstantTimeCompare(computedReceivedApiKeyHash, apiKeyHash) == 1 {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusUnauthorized)
	})

	api := router.Group("/api/v1")

	vehicles := api.Group("/vehicles")
	{
		vehicles.GET("/", handler.GetVehicles)
	}

	locations := api.Group("/locations")
	{
		locations.GET("/", handler.GetLocations)
		locations.GET("/latest", handler.GetLatestLocations)
	}

	return &Controller{Handler: handler, Router: router, GetApiKeys: getApiKeys, ApiKeyHashes: apiKeyHashes}, nil
}

func (c *Controller) Run(port int16) error {
	err := c.Router.Run(":" + strconv.Itoa(int(port)))
	if err != nil {
		return fmt.Errorf("ошибка запуска API: %w", err)
	}
	return nil
}
