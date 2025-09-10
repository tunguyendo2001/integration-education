package excel

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"integration-education-db/internal/model"
	"integration-education-db/pkg/errors"

	"github.com/xuri/excelize/v2"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, data []byte) ([]model.GradeRow, error) {
	// Create file from bytes
	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	// Get the first worksheet
	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.ErrInvalidFileFormat
	}

	sheetName := sheets[0]
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 { // Header + at least one data row
		return nil, errors.ErrInvalidFileFormat
	}

	// Parse header to find column indices
	header := rows[0]
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Validate required columns
	requiredColumns := []string{"student_id", "subject", "class", "semester", "grade"}
	for _, col := range requiredColumns {
		if _, exists := columnMap[col]; !exists {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var grades []model.GradeRow
	for i, row := range rows[1:] { // Skip header
		if len(row) < len(requiredColumns) {
			continue // Skip incomplete rows
		}

		grade, err := p.parseRow(row, columnMap, i+2) // i+2 for actual row number
		if err != nil {
			return nil, fmt.Errorf("error parsing row %d: %w", i+2, err)
		}

		grades = append(grades, *grade)
	}

	return grades, nil
}

func (p *Parser) parseRow(row []string, columnMap map[string]int, rowNum int) (*model.GradeRow, error) {
	getValue := func(colName string) string {
		if idx, exists := columnMap[colName]; exists && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	studentID := getValue("student_id")
	if studentID == "" {
		return nil, fmt.Errorf("student_id is required")
	}

	subject := getValue("subject")
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}

	class := getValue("class")
	if class == "" {
		return nil, fmt.Errorf("class is required")
	}

	semester := getValue("semester")
	if semester == "" {
		return nil, fmt.Errorf("semester is required")
	}

	gradeStr := getValue("grade")
	if gradeStr == "" {
		return nil, fmt.Errorf("grade is required")
	}

	grade, err := strconv.ParseFloat(gradeStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid grade value: %s", gradeStr)
	}

	return &model.GradeRow{
		StudentID: studentID,
		Subject:   subject,
		Class:     class,
		Semester:  semester,
		Grade:     grade,
	}, nil
}
