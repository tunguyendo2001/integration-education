CREATE DATABASE IF NOT EXISTS education_local CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE education_local;

-- ======================
-- 1. TEACHERS
-- ======================
CREATE TABLE teachers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ======================
-- 2. CLASSES
-- ======================
CREATE TABLE classes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    grade INT NOT NULL CHECK (grade BETWEEN 1 AND 12),
    school_year VARCHAR(20) NOT NULL,
    semester INT NOT NULL CHECK (semester BETWEEN 1 AND 2),
    is_active BOOLEAN DEFAULT TRUE
);

-- ======================
-- 3. SUBJECTS
-- ======================
CREATE TABLE subjects (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- ======================
-- 4. STUDENTS
-- ======================
CREATE TABLE students (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    gender ENUM('Nam', 'Nữ') NOT NULL,
    birth_date DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ======================
-- 5. TEACHER - CLASS ASSIGNMENTS
-- ======================
CREATE TABLE teacher_classes (
    teacher_id INT NOT NULL,
    class_id INT NOT NULL,
    PRIMARY KEY (teacher_id, class_id),
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);

-- ======================
-- 6. STUDENT - CLASS ASSIGNMENTS
-- ======================
CREATE TABLE student_class_assignments (
    student_id INT NOT NULL,
    class_id INT NOT NULL,
    PRIMARY KEY (student_id, class_id),
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
);
