INSERT INTO users (
    username,
    email,
    fullname,
    password_hash,
    base_salary,
    join_date,
    is_active,
    role,
    created_by,
    updated_by
)
SELECT 
    LOWER(REPLACE(fullname, ' ', '')) || gs.num,
    LOWER(REPLACE(fullname, ' ', '')) || gs.num || '@example.com',
    fullname,
    'pbkdf2_sha256$390000$customSalt123$qnXQ3SQKQ/1vXGPk5LylyVeQMY1Ws+pwhCJGJ+BhlD4=',
    ((2 + floor(random() * 9)) * 1000000)::int,
    date '2024-01-01' + (random() * 365)::int,
    true,
    CASE WHEN random() < 0.1 THEN 'admin'::user_roles ELSE 'employee'::user_roles END,
    'seeder',
    'seeder'
FROM (
    SELECT unnest(ARRAY[
        'Ayu Lestari', 'Budi Santoso', 'Citra Dewi', 'Dedi Pratama', 'Eka Saputra',
        'Fajar Nugroho', 'Gita Wulandari', 'Hadi Wijaya', 'Indah Permata', 'Joko Purnomo',
        'Kiki Ramadhani', 'Lina Marlina', 'Miko Febrian', 'Nina Kartika', 'Oscar Hidayat',
        'Putri Melati', 'Qory Alamsyah', 'Rian Hidayah', 'Sari Anggraeni', 'Teguh Prakoso',
        'Umi Kalsum', 'Vera Oktaviani', 'Wawan Setiawan', 'Xenia Natalia', 'Yusuf Maulana',
        'Zahra Ayuningtyas', 'Aditya Nugraha', 'Bella Salsabila', 'Cahyo Rahmat', 'Dina Safira',
        'Erlangga Mahendra', 'Farah Nabila', 'Gilang Ramadhan', 'Hilmi Ananda', 'Ika Yuliana',
        'Januar Rizki', 'Kirana Mawar', 'Lukman Hakim', 'Mega Dwi', 'Naufal Fadhil',
        'Okta Syahputra', 'Prita Ardhana', 'Qasim Ibrahim', 'Raisa Andriana', 'Surya Atmaja',
        'Tiara Kusuma', 'Ujang Sopandi', 'Vino G. Bastian', 'Winda Sari', 'Yana Mulyana'
    ]) AS fullname
) AS names
CROSS JOIN generate_series(1, 2) AS gs(num);


-- 

-- Attendance user 1 full masuk hari kerja Januari 2025 (diasumsikan Senin-Jumat)
INSERT INTO attendances (user_id, period, checkin_time, checkout_time, created_by, updated_by)
SELECT
  1,
  date,
  '09:00:00',
  '17:00:00',
  'seed',
  'seed'
FROM generate_series('2025-01-01'::date, '2025-01-31'::date, '1 day') AS date
WHERE EXTRACT(ISODOW FROM date) BETWEEN 1 AND 5;

-- Attendance user 2 cuma 10 hari kerja di Januari 2025, pilih tanggal ganjil hari kerja saja
INSERT INTO attendances (user_id, period, checkin_time, checkout_time, created_by, updated_by)
SELECT
  2,
  date,
  '09:00:00',
  '17:00:00',
  'seed',
  'seed'
FROM (
  SELECT date
  FROM generate_series('2025-01-01'::date, '2025-01-31'::date, '1 day') AS date
  WHERE EXTRACT(ISODOW FROM date) BETWEEN 1 AND 5
  AND EXTRACT(DAY FROM date)::int % 2 = 1
  LIMIT 10
) sub;

-- Overtime user 2, 3 tanggal berbeda
INSERT INTO overtimes (user_id, period, hours, reason, created_by, updated_by) VALUES
(2, '2025-01-10', 2, 'Project deadline', 'seed', 'seed'),
(2, '2025-01-15', 3, 'Urgent task', 'seed', 'seed'),
(2, '2025-01-20', 1, 'System maintenance', 'seed', 'seed');

-- Reimbursement user 1, 3 kali di Januari 2025
INSERT INTO reimbursements (user_id, period, amount, description, created_by, updated_by) VALUES
(1, '2025-01-05', 150000, 'Travel expense', 'seed', 'seed'),
(1, '2025-01-15', 200000, 'Meal allowance', 'seed', 'seed'),
(1, '2025-01-25', 100000, 'Office supplies', 'seed', 'seed');
