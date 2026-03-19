module github.com/lariv-in/p_nirmancampus_students

go 1.26.1

require (
	github.com/lariv-in/components v0.0.0
	github.com/lariv-in/getters v0.0.0
	github.com/lariv-in/lago v0.0.0
	github.com/lariv-in/p_students v0.0.0
	github.com/lariv-in/views v0.0.0
	gorm.io/gorm v1.31.1
)

replace (
	github.com/lariv-in/components => ../../components
	github.com/lariv-in/getters => ../../getters
	github.com/lariv-in/lago => ../../lago
	github.com/lariv-in/p_students => ../p_students
	github.com/lariv-in/views => ../../views
)

