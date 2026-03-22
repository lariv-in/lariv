module github.com/lariv-in/lago/p_nirmancampus_students

go 1.26.1

require (
	github.com/lariv-in/lago/components v0.0.0
	github.com/lariv-in/lago/getters v0.0.0
	github.com/lariv-in/lago/lago v0.0.0
	github.com/lariv-in/lago/p_students v0.0.0
	github.com/lariv-in/lago/views v0.0.0
	gorm.io/gorm v1.31.1
)

replace (
	github.com/lariv-in/lago/components => ../../components
	github.com/lariv-in/lago/getters => ../../getters
	github.com/lariv-in/lago/lago => ../../lago
	github.com/lariv-in/lago/p_students => ../p_students
	github.com/lariv-in/lago/views => ../../views
)

