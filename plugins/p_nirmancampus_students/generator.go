package p_nirmancampus_students

import (
	"math/rand"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_students"
	"gorm.io/gorm"
)

var studentCategories = []string{
	"General",
	"OBC",
	"SC",
	"ST",
	"",
}

var fathersNamePrefixes = []string{
	"Ravi",
	"Suresh",
	"Mahesh",
	"Ramesh",
	"Venkatesh",
	"Kumar",
	"Prakash",
	"Rajesh",
	"Mohan",
	"Raghav",
}

func randomAddress(r *rand.Rand) string {
	// Lightweight pseudo-data: enough to populate the fields without external faker deps.
	number := r.Intn(9999) + 1
	street := []string{"Main St", "Lake View", "Market Rd", "Park Ave", "Temple Rd"}[r.Intn(5)]
	city := []string{"Nirmancampus", "Hyderabad", "Pune", "Chennai", "Delhi"}[r.Intn(5)]
	pin := r.Intn(899999) + 100000
	return street + " " + itoa(number) + ", " + city + " - " + itoa(pin)
}

func randomFathersName(r *rand.Rand) string {
	if r.Intn(100) < 30 {
		// ~30% empty to match "optional" behavior in the source plugin.
		return ""
	}
	prefix := fathersNamePrefixes[r.Intn(len(fathersNamePrefixes))]
	suffix := r.Intn(999) + 1
	return prefix + " " + itoa(suffix)
}

func itoa(n int) string {
	// Avoid strconv import to keep this file compact; this is a trivial conversion.
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		d := n % 10
		buf = append(buf, byte('0'+d))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func init() {
	lago.RegistryGenerator.Register("students.NirmancampusStudentDetailsGenerator", lago.Generator{
		Create: func(db *gorm.DB) error {
			var students []p_students.Student
			if err := db.Find(&students).Error; err != nil {
				return err
			}

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for _, student := range students {
				fathersName := randomFathersName(r)
				category := studentCategories[r.Intn(len(studentCategories))]
				address := ""
				if r.Intn(100) < 60 {
					address = randomAddress(r)
				}

				err := db.Where(NirmancampusStudentDetails{StudentID: student.ID}).
					Assign(NirmancampusStudentDetails{
						FathersName: fathersName,
						Category:    category,
						Address:     address,
					}).
					FirstOrCreate(&NirmancampusStudentDetails{}).Error
				if err != nil {
					return err
				}
			}

			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1=1").Delete(&NirmancampusStudentDetails{}).Error
		},
	})
}
