CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'student',
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP,


    student_number VARCHAR(100),
    first_name VARCHAR(100) NOT NULL DEFAULT '',
    last_name VARCHAR(100) NOT NULL DEFAULT '',
    middle_name VARCHAR(100),
    course VARCHAR(100),
    year_level VARCHAR(50),
    section VARCHAR(50),
    department VARCHAR(100),
    college VARCHAR(100),
    contact_number VARCHAR(20),
    address TEXT,
    qr_code_data TEXT
); 






CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    event_name VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    course VARCHAR(100),
    department VARCHAR(100),
    college VARCHAR(100),
    created_by VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_college ON events(college);
CREATE INDEX idx_events_department ON events(department);
CREATE INDEX idx_events_course ON events(course);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_is_active ON events(is_active);




-- Add the new columns for event scope
ALTER TABLE events 
ADD COLUMN target_courses TEXT,
ADD COLUMN target_year_levels TEXT,
ADD COLUMN target_sections TEXT,
ADD COLUMN qr_code_type VARCHAR(100);

-- Remove the old course column since we're using target_courses instead
ALTER TABLE events DROP COLUMN IF EXISTS course;

-- Add index for the new columns for better performance
CREATE INDEX IF NOT EXISTS idx_events_qr_code_type ON events(qr_code_type);