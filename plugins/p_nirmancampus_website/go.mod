module github.com/lariv-in/lago/plugins/p_nirmancampus_website

go 1.26.1

require (
	github.com/lariv-in/lago/plugins/p_nirmancampus_programs v0.0.0
	gorm.io/gorm v1.31.1
	maragu.dev/gomponents v1.2.0
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/lariv-in/lago/plugins/p_nirmancampus_programs => ../p_nirmancampus_programs
