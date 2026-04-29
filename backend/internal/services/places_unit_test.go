package services

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HaversineMeters реплицирует SQL-формулу для проверки кроссбазового
// fallback-кода в SQL-запросе Nearby.
func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(lat2 - lat1)
	dLon := rad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

func TestHaversineKnownDistances(t *testing.T) {
	tests := []struct {
		name                   string
		lat1, lon1, lat2, lon2 float64
		expectedMeters         float64
		tolerance              float64
	}{
		{"Москва — Санкт-Петербург (~635 км)", 55.7558, 37.6173, 59.9343, 30.3351, 635000, 5000},
		{"Парк Горького — Зарядье (~3 км)", 55.7298, 37.6019, 55.7515, 37.6294, 3000, 500},
		{"Один и тот же пункт", 55.7558, 37.6173, 55.7558, 37.6173, 0, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := HaversineMeters(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			assert.InDelta(t, tt.expectedMeters, d, tt.tolerance,
				"расхождение более чем на %.0f м", tt.tolerance)
		})
	}
}

func TestPlaceFilterDefaults(t *testing.T) {
	f := PlaceFilter{}
	assert.Zero(t, f.NoiseMax)
	assert.Zero(t, f.LightMax)
	assert.Zero(t, f.CrowdMax)
}
