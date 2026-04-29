package handlers

import (
	"net/http"
	"strconv"

	"github.com/atakhanov/sensory-navigator/backend/internal/middleware"
	"github.com/atakhanov/sensory-navigator/backend/internal/models"
	"github.com/atakhanov/sensory-navigator/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type PlacesHandler struct {
	places *services.PlaceService
}

func NewPlacesHandler(s *services.PlaceService) *PlacesHandler {
	return &PlacesHandler{places: s}
}

// List   — GET /api/places
// Search — GET /api/places/search?q=...
// Объединённый эндпоинт, поддерживает фильтры:
//
//	?category_id=...&q=...&noise_max=3&light_max=3&crowd_max=3&smell_max=3&visual_max=3
func (h *PlacesHandler) List(c *gin.Context) {
	q := c.Query("q")
	filter := services.PlaceFilter{
		CategoryID: parseUint(c.Query("category_id")),
		Search:     q,
		NoiseMax:   parseInt(c.Query("noise_max")),
		LightMax:   parseInt(c.Query("light_max")),
		CrowdMax:   parseInt(c.Query("crowd_max")),
		SmellMax:   parseInt(c.Query("smell_max")),
		VisualMax:  parseInt(c.Query("visual_max")),
		Limit:      parseInt(c.Query("limit")),
		Offset:     parseInt(c.Query("offset")),
	}
	places, err := h.places.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": places, "count": len(places)})
}

// Get — GET /api/places/:id
func (h *PlacesHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}
	place, err := h.places.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "место не найдено"})
		return
	}
	c.JSON(http.StatusOK, place)
}

type createPlaceReq struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Address     string  `json:"address" binding:"max=500"`
	Description string  `json:"description"`
	CategoryID  uint64  `json:"category_id" binding:"required"`
	Latitude    float64 `json:"latitude" binding:"required,latitude"`
	Longitude   float64 `json:"longitude" binding:"required,longitude"`
}

// Create — POST /api/places (требует JWT)
func (h *PlacesHandler) Create(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)

	var req createPlaceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &models.Place{
		Name:        req.Name,
		Address:     req.Address,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		CreatedByID: uid,
	}
	if err := h.places.Create(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// Update — PUT /api/places/:id (требует JWT)
func (h *PlacesHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}
	place, err := h.places.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "место не найдено"})
		return
	}

	var req createPlaceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	place.Name = req.Name
	place.Address = req.Address
	place.Description = req.Description
	place.CategoryID = req.CategoryID
	place.Latitude = req.Latitude
	place.Longitude = req.Longitude
	if err := h.places.Update(place); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, place)
}

// Delete — DELETE /api/places/:id (требует JWT)
func (h *PlacesHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}
	if err := h.places.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Nearby — GET /api/places/nearby?lat=..&lon=..&radius=..(метры)
func (h *PlacesHandler) Nearby(c *gin.Context) {
	lat, err1 := strconv.ParseFloat(c.Query("lat"), 64)
	lon, err2 := strconv.ParseFloat(c.Query("lon"), 64)
	radius := parseInt(c.DefaultQuery("radius", "1000"))
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходимы параметры lat и lon"})
		return
	}
	places, err := h.places.Nearby(lat, lon, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": places, "count": len(places)})
}

// Categories — GET /api/categories
func (h *PlacesHandler) Categories(c *gin.Context) {
	cats, err := h.places.Categories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": cats, "count": len(cats)})
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseUint(s string) uint64 {
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}