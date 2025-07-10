package helpers

import (
	"math"
	"time"
)

const (
	NOT_CHECKED_IN = "not_checked_in"
	CHECKED_IN     = "checked_in"
	CHECKED_OUT    = "checked_out"
)

// leave status
const (
	PENDING_LEAVE  = 1
	APPROVED_LEAVE = 2
	REJECTED_LEAVE = 3
)

func IsLeaveDateValid(dates []string) (bool, string) {
	today := time.Now().Truncate(24 * time.Hour)

	for _, d := range dates {
		parsedDate, err := time.Parse("2006-01-02", d)
		if err != nil {
			return false, "Format tanggal tidak valid: " + d
		}
		if !parsedDate.After(today) {
			return false, "Tanggal cuti tidak boleh hari ini atau tanggal yang sudah lewat: " + d
		}
	}
	return true, ""
}

func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Meter
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}