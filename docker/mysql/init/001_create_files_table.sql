CREATE DATABASE IF NOT EXISTS education_local CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE education_local;

-- =====================================================
-- 1. TEACHERS (Data từ server)
-- =====================================================
CREATE TABLE teachers (
    id BIGINT PRIMARY KEY,  -- Dùng ID từ server, KHÔNG auto_increment
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255),
    gender ENUM('MEN', 'WOMEN'),
    hometown VARCHAR(255),
    birthday DATE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Metadata để tracking
    synced_from_server_at DATETIME,  -- Lần cuối pull từ server
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_email (email),
    INDEX idx_username (username),
    INDEX idx_active (is_active)
) COMMENT = 'Teachers data pulled from server - ID from server';

-- =====================================================
-- 2. STUDENTS (Data từ server)
-- =====================================================
CREATE TABLE students (
    id BIGINT PRIMARY KEY,  -- Dùng ID từ server, KHÔNG auto_increment
    name VARCHAR(100) NOT NULL,
    gender ENUM('Nam', 'Nữ') NOT NULL,
    hometown VARCHAR(255),
    birth_date DATE,
    
    -- Metadata để tracking
    synced_from_server_at DATETIME,  -- Lần cuối pull từ server
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_gender (gender)
) COMMENT = 'Students data pulled from server - ID from server';

-- =====================================================
-- 3. SUBJECTS
-- =====================================================
CREATE TABLE subjects (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- =====================================================
-- 4. CLASSES
-- =====================================================
CREATE TABLE classes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    grade INT NOT NULL CHECK (grade BETWEEN 1 AND 12),
    school_year VARCHAR(20) NOT NULL,
    semester INT NOT NULL CHECK (semester BETWEEN 1 AND 2),
    subject VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_year_semester (school_year, semester),
    INDEX idx_active (is_active)
);

-- =====================================================
-- 5. TEACHER - CLASS ASSIGNMENTS
-- =====================================================
CREATE TABLE teacher_classes (
    teacher_id BIGINT NOT NULL,  -- FK tới teachers.id (server ID)
    class_id INT NOT NULL,
    subject VARCHAR(100),
    PRIMARY KEY (teacher_id, class_id),
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    
    INDEX idx_teacher (teacher_id),
    INDEX idx_class (class_id)
);

-- =====================================================
-- 6. STUDENT - CLASS ASSIGNMENTS
-- =====================================================
CREATE TABLE student_class_assignments (
    student_id BIGINT NOT NULL,  -- FK tới students.id (server ID)
    class_id INT NOT NULL,
    PRIMARY KEY (student_id, class_id),
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    
    INDEX idx_student (student_id),
    INDEX idx_class (class_id)
);

-- =====================================================
-- 7. SCORES (Dữ liệu nhập ở local, sẽ push lên server)
-- =====================================================
CREATE TABLE scores (
    id INT AUTO_INCREMENT PRIMARY KEY,
    student_id BIGINT NOT NULL,   -- Dùng server student ID
    class_id INT NOT NULL,
    teacher_id BIGINT NOT NULL,   -- Dùng server teacher ID
    subject_id INT NOT NULL,
    school_year VARCHAR(20) NOT NULL,
    semester INT NOT NULL CHECK (semester BETWEEN 1 AND 2),
    
    -- Các loại điểm
    score_15min FLOAT CHECK (score_15min BETWEEN 0 AND 10),
    score_45min FLOAT CHECK (score_45min BETWEEN 0 AND 10),
    final_exam FLOAT CHECK (final_exam BETWEEN 0 AND 10),
    
    -- Điểm trung bình tự động tính
    avg_semester FLOAT GENERATED ALWAYS AS (
        ROUND((score_15min + score_45min*2 + final_exam*3)/6, 2)
    ) STORED,
    
    comment TEXT,
    
    -- Trạng thái đồng bộ
    is_synced BOOLEAN DEFAULT FALSE,  -- Đã push lên server chưa
    synced_at DATETIME NULL,  -- Thời điểm sync thành công
    sync_error TEXT NULL,  -- Lỗi nếu sync failed
    server_score_id VARCHAR(500) NULL,  -- ID của score trên server sau khi sync
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE,
    
    INDEX idx_student (student_id),
    INDEX idx_class (class_id),
    INDEX idx_teacher (teacher_id),
    INDEX idx_subject (subject_id),
    INDEX idx_sync_status (is_synced),
    INDEX idx_year_semester (school_year, semester),
    
    -- Đảm bảo không trùng điểm
    UNIQUE KEY uniq_score (student_id, teacher_id, subject_id, semester, school_year)
) COMMENT = 'Staging scores data, waiting to sync to server';

-- =====================================================
-- 8. SYNC LOG (Tracking sync operations)
-- =====================================================
CREATE TABLE sync_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sync_type ENUM('pull_teachers', 'pull_students', 'push_scores') NOT NULL,
    status ENUM('success', 'failed', 'in_progress') NOT NULL,
    records_count INT DEFAULT 0,
    error_message TEXT,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    
    INDEX idx_type (sync_type),
    INDEX idx_status (status),
    INDEX idx_started (started_at)
) COMMENT = 'Sync log';
