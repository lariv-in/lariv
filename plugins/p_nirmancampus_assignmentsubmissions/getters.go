package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
)

const assignmentSubmissionsEnvironmentSessionKey = "assignmentsubmissions_session"

func assignmentSubmissionsSessionEnvironmentDefault(ctx context.Context) (uint, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return 0, nil
	}
	id, err := sessions.DefaultAdmissionSessionID(db)
	if err != nil {
		return 0, nil
	}
	return id, nil
}

func bulkAcademicRecordStudentLineGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		rec, ok := bulkAcademicRecordFromContext(ctx)
		if !ok || rec.ID == 0 {
			return "", nil
		}
		return fmt.Sprintf("%s — %s", rec.Student.Name, rec.Student.StudentNo), nil
	}
}

// academicRecordForInputForeignKey returns a preloaded [p_nirmancampus_academicrecords.AcademicRecord] from
// request context: list filter layer ([listFilterAcademicRecordContextKey]), [LayerDetail] on
// [AssignmentSubmission] (assignmentsubmissions.detail / UpdateView), or $in from the bulk modal layer
// (merged [AcademicRecord] + [AcademicRecordID]).
func academicRecordForInputForeignKey() getters.Getter[p_nirmancampus_academicrecords.AcademicRecord] {
	return func(ctx context.Context) (p_nirmancampus_academicrecords.AcademicRecord, error) {
		var zero p_nirmancampus_academicrecords.AcademicRecord
		if ar, ok := ctx.Value(listFilterAcademicRecordContextKey).(p_nirmancampus_academicrecords.AcademicRecord); ok && ar.ID != 0 {
			return ar, nil
		}
		sub, err := getters.Key[AssignmentSubmission]("assignmentsubmission")(ctx)
		if err == nil && sub.ID != 0 && sub.AcademicRecordID != 0 && sub.AcademicRecord.ID != 0 && sub.AcademicRecord.ID == sub.AcademicRecordID {
			return sub.AcademicRecord, nil
		}
		if ar, err := getters.Key[p_nirmancampus_academicrecords.AcademicRecord]("$in.AcademicRecord")(ctx); err == nil && ar.ID != 0 {
			return ar, nil
		}
		return zero, nil
	}
}
