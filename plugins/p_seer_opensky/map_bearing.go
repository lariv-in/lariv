package p_seer_opensky

import "math"

// initialBearingDeg returns the initial (forward) bearing from (lat1,lng1) to (lat2,lng2) in WGS84,
// in degrees clockwise from true north, range [0, 360).
func initialBearingDeg(lat1, lng1, lat2, lng2 float64) float64 {
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180
	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	θ := math.Atan2(y, x)
	deg := θ * 180 / math.Pi
	return math.Mod(deg+360, 360)
}
