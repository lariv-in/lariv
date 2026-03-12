package p_totschool_appointments

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

var APPOINTMENT_NAMES = []string{
	"Initial Consultation",
	"Follow-up Meeting",
	"Project Kickoff",
	"Strategy Session",
	"Quarterly Review",
	"Annual Check-in",
	"Onboarding Session",
	"Technical Interview",
	"Design Review",
	"Stakeholder Meeting",
	"Brainstorming Session",
	"Sprint Planning",
	"Retrospective",
	"Client Presentation",
	"Vendor Negotiation",
}

var LOCATIONS = []string{
	"Conference Room A (Headquarters)",
	"Conference Room B (Headquarters)",
	"Virtual (Zoom Link: https://zoom.us/j/123456789)",
	"Virtual (Google Meet: https://meet.google.com/abc-defg-hij)",
	"Client Office (Downtown)",
	"Client Office (Northside)",
	"Coffee Shop (Main St.)",
	"Co-working Space (Desk 12)",
	"Branch Office (West)",
	"Branch Office (East)",
}

var TEMPLATES = []LetterTemplate{
	{
		Name:    "Standard Appointment Confirmation",
		Content: "Dear {Client Name},\n\nWe are writing to confirm your appointment for '{Appointment Name}' scheduled on {Appointment Date} at {Appointment Time}.\n\nThe meeting will take place at: {Location}.\n\nIf you have any questions or need to reschedule, please contact us at {Phone Number}.\n\nAdditional details:\n{Remarks}\n\nWe look forward to seeing you.\n\nBest regards,\n{Your Name}",
	},
	{
		Name:    "Virtual Meeting Details",
		Content: "Hello {Client Name},\n\n This is a reminder for our upcoming virtual meeting: '{Appointment Name}'.\n\nDate: {Appointment Date}\nTime: {Appointment Time}\nLocation/Link: {Location}\n\nPlease ensure you have a stable internet connection. If you encounter any technical issues joining, you can reach us at {Phone Number}.\n\nMeeting agenda/notes:\n{Remarks}\n\nBest,\n{Your Name}",
	},
	{
		Name:    "Follow-up Consultation",
		Content: "Dear {Client Name},\n\nThank you for scheduling your follow-up consultation.\n\nYour appointment, '{Appointment Name}', is confirmed for {Appointment Date} at {Appointment Time}. We will meet at {Location}.\n\nPlease review any materials discussed in our previous session before this meeting.\n\nQuestions? Call us at {Phone Number}.\n\nNotes for this session:\n{Remarks}\n\nSincerely,\n{Your Name}",
	},
}

func GenerateAppointmentsForUser(db *gorm.DB, user p_users.User, count int) {
	now := time.Now()
	// Create templates if they don't exist
	for _, tmpl := range TEMPLATES {
		var existing LetterTemplate
		if err := db.Where("name = ?", tmpl.Name).First(&existing).Error; err != nil {
			db.Create(&tmpl)
		}
	}

	for i := 0; i < count; i++ {
		// Calculate a random datetime within the next 30 days, between 9 AM and 5 PM
		daysOffset := rand.Intn(30) + 1 // 1 to 30 days from now
		hoursOffset := rand.Intn(8) + 9 // 9 AM to 4 PM (inclusive)
		minutesOffset := (rand.Intn(4)) * 15 // 0, 15, 30, 45 minutes

		apptDate := time.Date(now.Year(), now.Month(), now.Day()+daysOffset, hoursOffset, minutesOffset, 0, 0, now.Location())

		// Ensure no overlapping appointments for this user
		for {
			var overlappingCount int64
			db.Model(&Appointment{}).Where(
				"created_by_id = ? AND datetime = ?",
				user.ID, apptDate,
			).Count(&overlappingCount)

			if overlappingCount == 0 {
				break
			}
			// If overlap, shift by 30 minutes
			apptDate = apptDate.Add(30 * time.Minute)
		}

		name := APPOINTMENT_NAMES[rand.Intn(len(APPOINTMENT_NAMES))]
		location := LOCATIONS[rand.Intn(len(LOCATIONS))]

		// Random phone number (US format)
		phone := fmt.Sprintf("(%03d) %03d-%04d", rand.Intn(800)+200, rand.Intn(900)+100, rand.Intn(10000))

		remarks := ""
		if rand.Float64() > 0.5 {
			remarks = "Please review the attached documents before the meeting."
		}

		extraInfo := ""
		if rand.Float64() > 0.7 {
			extraInfo = "Client prefers formal tone. Mention their recent project."
		}

		appointment := Appointment{
			CreatedByID: user.ID,
			Name:        name,
			Location:    location,
			Datetime:    apptDate,
			Phone:       phone,
			Remarks:     remarks,
			ExtraInfo:   extraInfo,
		}
		db.Create(&appointment)
	}
}

func init() {
	lago.RegistryGenerator.Register("appointments.Generator", lago.Generator{
		Name:        "Appointments",
		Description: "Generates sample appointments and letter templates",
		Create: func(db *gorm.DB) error {
			var users []p_users.User
			if err := db.Find(&users).Error; err != nil {
				return err
			}

			// Generate appointments for each user
			for _, user := range users {
				// Base number of appointments + some randomness
				count := 10 + rand.Intn(15)
				GenerateAppointmentsForUser(db, user, count)
			}
			return nil
		},
		Remove: func(db *gorm.DB) error {
			if err := db.Unscoped().Where("1=1").Delete(&Appointment{}).Error; err != nil {
				return err
			}
			if err := db.Unscoped().Where("1=1").Delete(&LetterTemplate{}).Error; err != nil {
				return err
			}
			return nil
		},
	})
}
