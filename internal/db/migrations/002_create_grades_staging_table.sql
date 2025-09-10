CREATE TABLE IF NOT EXISTS grades_staging (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    file_id BIGINT NOT NULL,
    student_id VARCHAR(50) NOT NULL,
    subject VARCHAR(100) NOT NULL,
    class VARCHAR(50) NOT NULL,
    semester VARCHAR(20) NOT NULL,
    grade DECIMAL(5,2) NOT NULL CHECK (grade >= 0 AND grade <= 100),
    status VARCHAR(20) NOT NULL DEFAULT 'READY',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_grades_staging_file_id (file_id),
    INDEX idx_grades_staging_status (status),
    INDEX idx_grades_staging_class_subject (class, subject),
    INDEX idx_grades_staging_student_id (student_id),
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
