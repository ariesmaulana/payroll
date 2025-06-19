
CREATE TYPE user_roles AS ENUM ('employee', 'admin');


CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    fullname VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    base_salary INTEGER NOT NULL,
    join_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    role user_roles NOT NULL DEFAULT 'employee',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    updated_by VARCHAR(50)
);
-- For testing purpose
INSERT INTO users (username, email, fullname, password_hash, role, base_salary, join_date)
VALUES 
    ('testuser1', 'test1@example.com', 'User One', 'hashedpassword1' , 'admin', 100, '2025-01-01'),
    ('testuser2', 'test2@example.com', 'User Two', 'hashedpassword2', 'employee', 12121, '2025-02-01'),
    ('testuser3', 'test3@example.com', 'User Three', 'hashedpassword3', 'employee', 12121, '2025-03-01')
ON CONFLICT (email) DO NOTHING;
