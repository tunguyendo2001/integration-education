package excel

import (
	"context"

	"integration-education-db/internal/model"
)

type ParsingStrategy interface {
	Parse(ctx context.Context, data []byte) ([]model.GradeRow, error)
	Validate(ctx context.Context, grades []model.GradeRow) error
}

type ExcelStrategy struct {
	parser    *Parser
	validator *Validator
}

func NewExcelStrategy() ParsingStrategy {
	return &ExcelStrategy{
		parser:    NewParser(),
		validator: NewValidator(),
	}
}

func (s *ExcelStrategy) Parse(ctx context.Context, data []byte) ([]model.GradeRow, error) {
	return s.parser.Parse(ctx, data)
}

func (s *ExcelStrategy) Validate(ctx context.Context, grades []model.GradeRow) error {
	return s.validator.Validate(ctx, grades)
}
