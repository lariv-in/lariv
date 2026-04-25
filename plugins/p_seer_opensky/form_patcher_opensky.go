package p_seer_opensky

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/views"
)

// openSkyFormPatcher normalizes form values for [OpenSkyState] optional pointers and [SensorsText].
type openSkyFormPatcher struct{}

func (openSkyFormPatcher) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	// WGS84 lon/lat (non-pointer on model; 0,0 = unknown)
	parseFloat64FieldToZero(formData, "Longitude")
	parseFloat64FieldToZero(formData, "Latitude")

	// Pointers: empty or "-" → nil; otherwise parse
	ptrFloats := []string{
		"BaroAltitude", "Velocity", "TrueTrack", "VerticalRate", "GeoAltitude",
	}
	for _, k := range ptrFloats {
		stripToNilPtrFloat(formData, k)
	}
	if v, ok := formData["TimePosition"]; ok {
		s, isStr := v.(string)
		if isStr {
			s = strings.TrimSpace(s)
			if s == "" {
				delete(formData, "TimePosition")
			} else {
				if n, err := strconv.ParseInt(s, 10, 64); err == nil {
					formData["TimePosition"] = &n
				}
			}
		}
	}
	if _, ok := formData["PositionSource"]; ok {
		stripToNilPtrInt(formData, "PositionSource")
	}
	if _, ok := formData["Category"]; ok {
		stripToNilPtrInt(formData, "Category")
	}
	parseOptBool := func(k string) {
		v, ok := formData[k]
		if !ok {
			return
		}
		s, isStr := v.(string)
		if isStr {
			s = strings.TrimSpace(strings.ToLower(s))
			if s == "" {
				delete(formData, k)
				return
			}
			b := s == "true" || s == "1" || s == "yes"
			formData[k] = &b
		}
	}
	parseOptBool("OnGround")
	parseOptBool("SPI")
	return formData, formErrors
}

func parseFloat64FieldToZero(formData map[string]any, k string) {
	v, ok := formData[k]
	if !ok {
		return
	}
	if f, isFloat64 := v.(float64); isFloat64 {
		formData[k] = f
		return
	}
	s, isStr := v.(string)
	if !isStr {
		return
	}
	s = strings.TrimSpace(s)
	if s == "" {
		formData[k] = float64(0)
		return
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return
	}
	formData[k] = f
}

func stripToNilPtrFloat(formData map[string]any, k string) {
	v, ok := formData[k]
	if !ok {
		return
	}
	s, isStr := v.(string)
	if !isStr {
		return
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "—" {
		delete(formData, k)
		return
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return
	}
	formData[k] = &f
}

func stripToNilPtrInt(formData map[string]any, k string) {
	v, ok := formData[k]
	if !ok {
		return
	}
	s, isStr := v.(string)
	if !isStr {
		return
	}
	s = strings.TrimSpace(s)
	if s == "" {
		delete(formData, k)
		return
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	formData[k] = &n
}

var openSkyFormPatchers = views.FormPatchers{
	{Key: "p_seer_opensky.form", Value: openSkyFormPatcher{}},
}
