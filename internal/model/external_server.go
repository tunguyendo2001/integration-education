package model

import "time"

// ServerTeacher represents teacher data from server API
type ServerTeacher struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Gender       string  `json:"gender"`
	Hometown     string  `json:"hometown"`
	Birthday     string  `json:"birthday"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"passwordHash"`
	Email        string  `json:"email"`
	IsActive     bool    `json:"isActive"`
	LastLogin    *string `json:"lastLogin"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// ServerStudent represents student data from server API
type ServerStudent struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	Hometown  string `json:"hometown"`
	Birthday  string `json:"birthday"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ServerClass represents class data from server API
type ServerClass struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ClassName    string `json:"className"`
	GradeLevel   int    `json:"gradeLevel"`
	AcademicYear int    `json:"academicYear"`
	Semester     string `json:"semester"`
	Subject      string `json:"subject"`
	IsActive     bool   `json:"isActive"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// SemesterPermissionRequest represents the request for checking semester permission
type SemesterPermissionRequest struct {
	Semester  string `json:"semester"`
	Year      int    `json:"year"`
	ClassName string `json:"className"`
}

// SemesterSchedule represents schedule info from server
type SemesterSchedule struct {
	ID               int64     `json:"id"`
	ScheduleName     string    `json:"scheduleName"`
	Semester         string    `json:"semester"`
	Year             int       `json:"year"`
	ClassName        string    `json:"className"`
	StartDateTime    time.Time `json:"startDateTime"`
	EndDateTime      time.Time `json:"endDateTime"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	CreatedBy        string    `json:"createdBy"`
	Active           bool      `json:"active"`
	Locked           bool      `json:"locked"`
	CurrentlyAllowed bool      `json:"currentlyAllowed"`
}

// SemesterPermissionResponse represents the response from check-permission API
type SemesterPermissionResponse struct {
	CurrentTime    time.Time   `json:"currentTime"`
	IsAllowed      bool        `json:"isAllowed"`
	Year           int         `json:"year"`
	Semester       string      `json:"semester"`
	ClassName      string      `json:"className"`
	ActiveSchedule interface{} `json:"activeSchedule"` // Can be string or SemesterSchedule
}

// GetActiveSchedule safely extracts the active schedule
func (r *SemesterPermissionResponse) GetActiveSchedule() *SemesterSchedule {
	if schedule, ok := r.ActiveSchedule.(map[string]interface{}); ok {
		// Convert map to SemesterSchedule
		if id, ok := schedule["id"].(float64); ok {
			return &SemesterSchedule{
				ID:               int64(id),
				ScheduleName:     getStringField(schedule, "scheduleName"),
				Semester:         getStringField(schedule, "semester"),
				Year:             getIntField(schedule, "year"),
				ClassName:        getStringField(schedule, "className"),
				Active:           getBoolField(schedule, "active"),
				Locked:           getBoolField(schedule, "locked"),
				CurrentlyAllowed: getBoolField(schedule, "currentlyAllowed"),
			}
		}
	}
	return nil
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getIntField(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
