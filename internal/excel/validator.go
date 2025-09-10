package excel

import (
	"context"
	"regexp"

	"integration-education-db/internal/model"
	"integration-education-db/pkg/errors"
)

type Validator struct {
	studentIDRegex *regexp.Regexp
}

func NewValidator() *Validator {
	return &Validator{
		studentIDRegex: regexp.MustCompile(`^[A-Z0-9]{6,20}$`),
	}
}

func (v *Validator) Validate(ctx context.Context, grades []model.GradeRow) error {
	if len(grades) == 0 {
		return errors.ErrSchemaValidation
	}

	for i, grade := range grades {
		if err := v.validateGrade(grade, i+1); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateGrade(grade model.GradeRow, rowNum int) error {
	// Validate student ID format
	if !v.studentIDRegex.MatchString(grade.StudentID) {
		return errors.ValidationError{
			Field:   "student_id",
			Value:   grade.StudentID,
			Message: "must be 6-20 alphanumeric characters",
		}
	}

	// Validate grade range
	if grade.Grade < 0 || grade.Grade > 10 {
		return errors.ValidationError{
			Field:   "grade",
			Value:   grade.Grade,
			Message: "must be between 0 and 10",
		}
	}

	// Validate subject not empty
	if len(grade.Subject) == 0 || len(grade.Subject) > 100 {
		return errors.ValidationError{
			Field:   "subject",
			Value:   grade.Subject,
			Message: "subject cannot be empty",
		}
	}

	// Validate class not empty
	if len(grade.Class) == 0 || len(grade.Class) > 50 {
		return errors.ValidationError{
			Field:   "class",
			Value:   grade.Class,
			Message: "class cannot be empty",
		}
	}

	// Validate semester not empty
	if len(grade.Semester) == 0 || len(grade.Semester) > 20 {
		return errors.ValidationError{
			Field:   "semester",
			Value:   grade.Semester,
			Message: "semester cannot be empty",
		}
	}

	return nil
}
