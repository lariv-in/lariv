package p_nirmancampus_examregistrations

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type examRegistrationScopeByRole struct{}

func (examRegistrationScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[ExamRegistration]) gorm.ChainInterface[ExamRegistration] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "ExamRegistrationScopeByRole")

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("ExamRegistrationScopeByRole: db from context", "error", err)
		panic("ExamRegistrationScopeByRole: " + err.Error())
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return query.Where("1 = 0")
		}
		studentSub := db.Model(&p_nirmancampus_students.Student{}).
			Select("id").
			Where("email = ?", email)
		academicRecordSub := db.Table("academic_records").
			Select("id").
			Where("student_id IN (?)", studentSub)
		return query.Where("academic_record_id IN (?)", academicRecordSub)
	default:
		return query.Where("1 = 0")
	}
}

// ExamRegistrationScopeByRole restricts queries by role (same rules as assignment submissions).
var ExamRegistrationScopeByRole views.QueryPatcher[ExamRegistration] = examRegistrationScopeByRole{}

type examRegistrationListSessionFilter struct{}

func selectedExamRegistrationsSessionFilter(db *gorm.DB, ctx context.Context) (id uint, restrict bool) {
	envMap, ok := ctx.Value("$environment").(map[string]string)
	if !ok {
		id, err := sessions.DefaultAdmissionSessionID(db)
		if err != nil {
			slog.Error("selectedExamRegistrationsSessionFilter: no default session", "error", err)
			return 0, false
		}
		return id, true
	}
	raw, inMap := envMap[examRegistrationsEnvironmentSessionKey]
	if !inMap {
		id, err := sessions.DefaultAdmissionSessionID(db)
		if err != nil {
			slog.Error("selectedExamRegistrationsSessionFilter: no default session", "error", err)
			return 0, false
		}
		return id, true
	}
	if strings.TrimSpace(raw) == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed == 0 {
		slog.Error("selectedExamRegistrationsSessionFilter: invalid session id in environment", "raw", raw, "error", err)
		id, defErr := sessions.DefaultAdmissionSessionID(db)
		if defErr != nil {
			return 0, false
		}
		return id, true
	}
	return uint(parsed), true
}

func (examRegistrationListSessionFilter) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[ExamRegistration]) gorm.ChainInterface[ExamRegistration] {
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("examRegistrationListSessionFilter: db from context", "error", err)
		return query
	}
	id, restrict := selectedExamRegistrationsSessionFilter(db, r.Context())
	if !restrict {
		return query
	}
	academicRecordSub := db.Model(&p_nirmancampus_academicrecords.AcademicRecord{}).
		Select("id").
		Where("session_id = ?", id)
	return query.Where("academic_record_id IN (?)", academicRecordSub)
}

var ExamRegistrationListSessionFilter views.QueryPatcher[ExamRegistration] = examRegistrationListSessionFilter{}

type examRegistrationListOrder struct{}

func (examRegistrationListOrder) Patch(_ views.View, _ *http.Request, query gorm.ChainInterface[ExamRegistration]) gorm.ChainInterface[ExamRegistration] {
	return query.Order("created_at DESC").Order("id DESC")
}

// ExamRegistrationListOrder is the default sort for the list view (newest first).
var ExamRegistrationListOrder views.QueryPatcher[ExamRegistration] = examRegistrationListOrder{}
