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

CREATE TABLE IF NOT EXISTS attendances (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    period DATE NOT NULL,
    checkin_time TIME,
    checkout_time TIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    updated_by VARCHAR(50),

    CONSTRAINT unique_attendance_per_period UNIQUE (user_id, period)
);

CREATE TABLE IF NOT EXISTS overtimes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    period DATE NOT NULL,
    hours INT NOT NULL CHECK (hours > 0 AND hours <= 3),
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    updated_by VARCHAR(50),
    CONSTRAINT unique_overtime_per_day UNIQUE (user_id, period)
);

CREATE TABLE IF NOT EXISTS reimbursements (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    period DATE NOT NULL,
    amount INT NOT NULL CHECK (amount > 0),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    updated_by VARCHAR(50),
    CONSTRAINT unique_reimbursement_per_day UNIQUE (user_id, period)
);

CREATE TABLE IF NOT EXISTS payrolls (
    id SERIAL PRIMARY KEY,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_attendance INT NOT NULL DEFAULT 0,
    total_overtime INT NOT NULL DEFAULT 0,
    total_reimbursement INT NOT NULL DEFAULT 0,
    total_salary INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    updated_by VARCHAR(50),
    CONSTRAINT unique_payroll_period UNIQUE (period_start, period_end)
);

CREATE TABLE IF NOT EXISTS payroll_items (
    id SERIAL PRIMARY KEY,
    payroll_id INT NOT NULL REFERENCES payrolls(id) ON DELETE CASCADE,
    user_id INT NOT NULL,
    attendance_count INT NOT NULL,
    overtime_hours INT NOT NULL,
    reimbursement_total INT NOT NULL,
    total_salary INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(50),
    updated_by VARCHAR(50),
    CONSTRAINT unique_user_payroll UNIQUE (payroll_id, user_id)
);
