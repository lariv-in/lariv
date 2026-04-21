package p_seer_gdelt

// gdeltValidLatLng returns false for (0,0) and coordinates outside WGS84 ranges.
func gdeltValidLatLng(lat, lng float64) bool {
	if lat == 0 && lng == 0 {
		return false
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return false
	}
	return true
}
