package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/tg-waqi-bot/pkg/waqi"
)

type restController struct {
	service waqi.Service
}

type getByGeoQuery struct {
	Lon float32 `form:"lon"`
	Lat float32 `form:"lat"`
}

// GetByGeo handles request /api/status/geo?lon=123&lat=456
func (ctrl *restController) GetByGeo(c *gin.Context) {
	var query getByGeoQuery
	err := c.BindQuery(&query)
	if err != nil {
		panic(err)
	}

	resp, err := ctrl.service.GetByGeo(query.Lat, query.Lon)
	if err != nil {
		panic(err)
	}

	c.JSON(200, resp)
}

// GetByCity handles request GET /api/status/city/:city
func (ctrl *restController) GetByCity(c *gin.Context) {
	city := c.Param("city")

	resp, err := ctrl.service.GetByCity(city)
	if err != nil {
		panic(err)
	}

	c.JSON(200, resp)
}

// GetByStation handles request GET /api/status/station/:id
func (ctrl *restController) GetByStation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		panic(err)
	}

	resp, err := ctrl.service.GetByStation(id)
	if err != nil {
		panic(err)
	}

	c.JSON(200, resp)
}
