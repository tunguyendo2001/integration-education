CREATE TABLE scores (
    id INT AUTO_INCREMENT PRIMARY KEY,
    student_id INT NOT NULL,
    class_id INT NOT NULL,
    teacher_id INT NOT NULL,
    subject_id INT NOT NULL,
    school_year VARCHAR(20) NOT NULL,
    semester INT NOT NULL CHECK (semester BETWEEN 1 AND 2),
    score_15min FLOAT CHECK (score_15min BETWEEN 0 AND 10),
    score_45min FLOAT CHECK (score_45min BETWEEN 0 AND 10),
    final_exam FLOAT CHECK (final_exam BETWEEN 0 AND 10),
    avg_semester FLOAT GENERATED ALWAYS AS (
        ROUND((score_15min + score_45min*2 + final_exam*3)/6, 2)
    ) STORED,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE,
    INDEX idx_scores_student (student_id),
    INDEX idx_scores_class (class_id),
    INDEX idx_scores_teacher (teacher_id),
    INDEX idx_scores_subject (subject_id)
);
