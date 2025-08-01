package api

import (
	"fmt"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	Handler *Handler
	Router  *gin.Engine
}

func NewController(handler *Handler) *Controller {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		AllowCredentials: false,
	}))

	vehicles := router.Group("/vehicles")
	{
		vehicles.GET("/", handler.GetVehicles)
	}

	locations := router.Group("/locations")
	{
		locations.GET("/", handler.GetLocations)
		locations.GET("/latest", handler.GetLatestLocations)
	}

	return &Controller{Handler: handler, Router: router}
}

func (c *Controller) Run(port int16) error {
	err := c.Router.Run(":" + strconv.Itoa(int(port)))
	if err != nil {
		logrus.Infof("API запущено на порту %d", port)
		return fmt.Errorf("ошибка запуска API: %w", err)
	}
	return nil
}
