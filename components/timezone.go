package components

import (
	"time"
)

var DefaultTimeZone *time.Location = time.FixedZone("Asia/Kolkata", int((time.Hour*5 + time.Minute*30).Seconds()))
