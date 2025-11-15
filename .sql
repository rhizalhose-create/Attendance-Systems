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

    -- College fields
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