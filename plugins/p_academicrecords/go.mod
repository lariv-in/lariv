module github.com/lariv-in/lago/plugins/p_academicrecords

go 1.26.1

require (
	github.com/lariv-in/lago/plugins/p_nirmancampus_students v0.0.0-00010101000000-000000000000
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/lariv-in/lago/plugins/p_nirmancampus_students => ../p_nirmancampus_students
