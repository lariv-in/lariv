package forms

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// reorderFormField moves a field one position up (earlier in sort_order order) or down.
func reorderFormField(db *gorm.DB, formID, fieldID uint, moveUp bool) error {
	fields, err := gorm.G[FormField](db).Where("form_id = ?", formID).Order("sort_order ASC, id ASC").Find(context.Background())
	if err != nil {
		return err
	}
	idx := -1
	for i := range fields {
		if fields[i].ID == fieldID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("forms: field %d not found for form %d", fieldID, formID)
	}
	if moveUp {
		if idx == 0 {
			return nil
		}
		return swapFieldSortOrder(db, &fields[idx], &fields[idx-1])
	}
	if idx >= len(fields)-1 {
		return nil
	}
	return swapFieldSortOrder(db, &fields[idx], &fields[idx+1])
}

func swapFieldSortOrder(db *gorm.DB, a, b *FormField) error {
	return db.Transaction(func(tx *gorm.DB) error {
		sa, sb := a.SortOrder, b.SortOrder
		if err := tx.Model(&FormField{}).Where("id = ?", a.ID).Update("sort_order", sb).Error; err != nil {
			return err
		}
		if err := tx.Model(&FormField{}).Where("id = ?", b.ID).Update("sort_order", sa).Error; err != nil {
			return err
		}
		return nil
	})
}
