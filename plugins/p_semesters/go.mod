module github.com/lariv-in/p_semesters

go 1.26.1

require (
	github.com/lariv-in/components v0.0.0
	github.com/lariv-in/getters v0.0.0
	github.com/lariv-in/lago v0.0.0
	github.com/lariv-in/p_users v0.0.0
	github.com/lariv-in/views v0.0.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/lariv-in/components => ../../components
	github.com/lariv-in/getters => ../../getters
	github.com/lariv-in/lago => ../../lago
	github.com/lariv-in/p_users => ../p_users
	github.com/lariv-in/views => ../../views
)
