package p_users

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const defaultPassword = "Pass1234#"

var indianFirstNames = []string{
	"Arjun", "Rahul", "Vikram", "Rajesh", "Amit",
	"Suresh", "Karthik", "Ravi", "Sandeep", "Prakash",
	"Naveen", "Deepak", "Manoj", "Sunil", "Anand",
	"Vijay", "Gaurav", "Rohit", "Sachin", "Aditya",
	"Kunal", "Ritesh", "Piyush", "Abhishek", "Vivek",
	"Rajiv", "Sanjay", "Dinesh", "Mukesh", "Harish",
	"Priya", "Anjali", "Neha", "Pooja", "Meera",
	"Divya", "Shweta", "Ritu", "Sneha", "Kavita",
	"Anita", "Deepika", "Rashmi", "Swati", "Jyoti",
	"Manisha", "Pallavi", "Rachana", "Shilpa", "Tanvi",
	"Aarti", "Bhavana", "Chitra", "Dipika", "Ekta",
	"Farah", "Geeta", "Hema", "Indira", "Jaya",
}

var indianLastNames = []string{
	"Sharma", "Patel", "Singh", "Kumar", "Gupta",
	"Verma", "Chopra", "Reddy", "Kapoor", "Malhotra",
	"Mehta", "Joshi", "Chauhan", "Agarwal", "Nair",
	"Menon", "Iyer", "Pillai", "Nayak", "Mishra",
	"Tiwari", "Bhat", "Desai", "Kaur", "Rao",
	"Khan", "Ali", "Hussain", "Khatri", "Gandhi",
	"Chatterjee", "Banerjee", "Mukherjee", "Sen", "Das",
	"Dutta", "Bose", "Ghosh", "Chakrabarti", "Basu",
	"Nambiar", "Kurup", "Warrier", "Namboothiri", "Panicker",
	"Thampi", "Nambissan", "Kartha", "Nair", "Menon",
	"Shetty", "Hegde", "Pai", "Kamath", "Bhat",
	"Rao", "Prabhu", "Kunder", "Kotian", "Salian",
}

// GetRandomIndianName returns a random combination of Indian first and last name.
func GetRandomIndianName() string {
	first := indianFirstNames[rand.Intn(len(indianFirstNames))]
	last := indianLastNames[rand.Intn(len(indianLastNames))]
	return first + " " + last
}

// GenerateRandomPhone generates a random Indian phone number in the format +91XXXXXXXXXX.
func GenerateRandomPhone() string {
	digits := make([]byte, 10)
	for i := range digits {
		digits[i] = byte('0' + rand.Intn(10))
	}
	return "+91" + string(digits)
}

// GenerateUser creates a user with realistic Indian data and a default password,
// assigned to the given role (created if it doesn't exist).
func GenerateUser(db *gorm.DB, roleName string) (*User, error) {
	name := GetRandomIndianName()

	userCount, err := gorm.G[User](db).Count(context.Background(), "*")
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s_%d", strings.ToLower(strings.ReplaceAll(name, " ", ".")), userCount+1)

	role := Role{Name: roleName}
	db.Where("name = ?", roleName).FirstOrCreate(&role)

	user := User{
		Name:     name,
		Email:    fmt.Sprintf("%s@school1.com", username),
		Phone:    GenerateRandomPhone(),
		Password: []byte(defaultPassword),
		RoleID:   role.ID,
	}
	if err := gorm.G[User](db).Create(context.Background(), &user); err != nil {
		return nil, err
	}
	return new(user), nil
}

// GenerateUserWithoutPassword creates a user with realistic Indian data but no password.
// This is faster because it skips password hashing.
func GenerateUserWithoutPassword(db *gorm.DB, roleName string) (*User, error) {
	name := GetRandomIndianName()

	userCount, err := gorm.G[User](db).Count(context.Background(), "*")
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s_%d", strings.ToLower(strings.ReplaceAll(name, " ", ".")), userCount+1)

	role := Role{Name: roleName}
	db.Where("name = ?", roleName).FirstOrCreate(&role)

	user := User{
		Name:   name,
		Email:  fmt.Sprintf("%s@school1.com", username),
		Phone:  GenerateRandomPhone(),
		RoleID: role.ID,
	}
	if err := gorm.G[User](db).Create(context.Background(), &user); err != nil {
		return nil, err
	}
	return new(user), nil
}

// CreateOverallSuperuser idempotently creates the system superuser (superadmin@lariv.in).
// Returns the existing user if already present.
func CreateOverallSuperuser(db *gorm.DB) (*User, error) {
	existing, err := gorm.G[User](db).Where("email = ?", "superadmin@lariv.in").First(context.Background())
	if err == nil {
		fmt.Println("Overall superuser already exists")
		return new(existing), nil
	}

	role := Role{Name: "superuser"}
	db.Where("name = ?", "superuser").FirstOrCreate(&role)

	user := User{
		Name:        "Super Admin",
		Email:       "superadmin@lariv.in",
		Password:    []byte(defaultPassword),
		IsSuperuser: true,
		RoleID:      role.ID,
	}
	if err := gorm.G[User](db).Create(context.Background(), &user); err != nil {
		return nil, err
	}
	fmt.Println("Created overall superuser")
	return new(user), nil
}

func init() {
	lago.RegistryGenerator.Register("users.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			_, err := CreateOverallSuperuser(db)
			return err
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Unscoped().Where("1=1").Delete(&User{}).Error; err != nil {
				return err
			}
			return nil
		},
	})
}
