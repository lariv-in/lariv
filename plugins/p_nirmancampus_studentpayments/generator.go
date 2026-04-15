package p_nirmancampus_studentpayments

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

func init() {
	methodKeys := registry.KeysFromPairs(PaymentMethodChoices)

	lago.RegistryGenerator.Register("studentpayments.Generator", lago.Generator{
		Create: func(db *gorm.DB) error {
			students, err := gorm.G[p_nirmancampus_students.Student](db).Order("id ASC").Find(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load students: %w", err)
			}
			if len(students) == 0 {
				return fmt.Errorf("need at least one student before generating payments")
			}
			if len(methodKeys) == 0 {
				return fmt.Errorf("payment method choices are empty")
			}

			today := time.Now().UTC()
			baseDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

			n := 0
			for si, st := range students {
				for j := 0; j < 2; j++ {
					idx := si*2 + j
					paid := baseDate.AddDate(0, 0, -(idx % 120))
					paidPtr := &paid
					method := methodKeys[idx%len(methodKeys)]
					amt := 1500.0 + float64(idx)*250.75
					p := Payment{
						StudentID:     st.ID,
						Amount:        amt,
						PaymentMethod: method,
						Remarks:       fmt.Sprintf("Generated payment %d for student %s", idx+1, st.StudentNo),
						TransactionID: fmt.Sprintf("GEN-%d-%d", st.ID, j),
						PaidAt:        paidPtr,
					}
					if err := gorm.G[Payment](db).Create(context.Background(), &p); err != nil {
						return fmt.Errorf("failed to create payment (student_id=%d): %w", st.ID, err)
					}
					n++
				}
			}

			fmt.Printf("Created %d payments (%d students × 2)\n", n, len(students))
			return nil
		},
		Remove: func(db *gorm.DB) error {
			return db.Unscoped().Where("1 = 1").Delete(&Payment{}).Error
		},
	})
}
