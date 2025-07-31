package api

import "github.com/gin-gonic/gin"

type Controller struct {
	Handler *Handler
}

func (c *Controller) New(handler *Handler, port int8) *Controller {
	router := gin.Default()

	vehicles := router.Group("/vehicles")
	{
		vehicles.GET(":provider_id", handler.GetVehicles)
	}

	router.Run(":" + string(rune(port)))

	return &Controller{Handler: handler}
}
