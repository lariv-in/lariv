module github.com/lariv-in/lago/p_courses_teachers

go 1.26.1

require (
	github.com/lariv-in/lago/components v0.0.0
	github.com/lariv-in/lago/getters v0.0.0
	github.com/lariv-in/lago/lago v0.0.0
	github.com/lariv-in/lago/p_courses v0.0.0
	github.com/lariv-in/lago/p_teachers v0.0.0
	github.com/lariv-in/lago/views v0.0.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/lariv-in/lago/components => ../../components
	github.com/lariv-in/lago/getters => ../../getters
	github.com/lariv-in/lago/lago => ../../lago
	github.com/lariv-in/lago/p_courses => ../p_courses
	github.com/lariv-in/lago/p_teachers => ../p_teachers
	github.com/lariv-in/lago/views => ../../views
)
