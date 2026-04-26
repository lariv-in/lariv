package p_nirmancampus_sessions

import "gorm.io/gorm"

func DefaultAdmissionSessionID(db *gorm.DB) (uint, error) {
	var active AdmissionSession
	err := db.Where("is_active = ?", true).Order(`"start" DESC`).First(&active).Error
	if err == nil && active.ID != 0 {
		return active.ID, nil
	}
	var latest AdmissionSession
	err = db.Order(`"start" DESC`).First(&latest).Error
	if err != nil {
		return 0, err
	}
	return latest.ID, nil
}
