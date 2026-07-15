-- Dealerships
INSERT INTO dealerships (id, name, address, opening_hours) VALUES
('d1', 'Saigon Auto', '123 Nguyen Hue, District 1, Ho Chi Minh City', '{"mon":"08:00-17:00","tue":"08:00-17:00","wed":"08:00-17:00","thu":"08:00-17:00","fri":"08:00-17:00","sat":"08:00-12:00","sun":"closed"}');
INSERT INTO dealerships (id, name, address, opening_hours) VALUES
('d2', 'Ha Noi Motors', '456 Le Duan, Hoan Kiem, Ha Noi', '{"mon":"08:00-17:00","tue":"08:00-17:00","wed":"08:00-17:00","thu":"08:00-17:00","fri":"08:00-17:00","sat":"08:00-16:00","sun":"closed"}');

-- Customers
INSERT INTO customers (id, name, email, phone) VALUES
('c1', 'Anh Tuan', 'anhtuan@example.com', '0901234567');
INSERT INTO customers (id, name, email, phone) VALUES
('c2', 'Chi Lan', 'chilan@example.com', '0907654321');
INSERT INTO customers (id, name, email, phone) VALUES
('c3', 'Bao Minh', 'baominh@example.com', '0901122334');

-- Vehicles
INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES
('v1', 'c1', 'VIN-TOYOTA-CAMRY-2023', 'Toyota', 'Camry', 2023);
INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES
('v2', 'c2', 'VIN-HONDA-CIVIC-2022', 'Honda', 'Civic', 2022);
INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES
('v3', 'c3', 'VIN-MAZDA-CX5-2024', 'Mazda', 'CX-5', 2024);

-- Service Types
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s1', 'Oil Change', 60, 'Full synthetic oil change with filter replacement');
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s2', 'Brake Replacement', 120, 'Front and rear brake pad replacement');
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s3', 'Engine Diagnostic', 90, 'Comprehensive engine diagnostic scan and report');
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s4', 'Tire Rotation', 30, 'Rotate and balance all four tires');

-- Technicians at d1 (Saigon)
INSERT INTO technicians (id, dealership_id, name) VALUES
('t1', 'd1', 'Minh');
INSERT INTO technicians (id, dealership_id, name) VALUES
('t2', 'd1', 'Hai');
INSERT INTO technicians (id, dealership_id, name) VALUES
('t3', 'd1', 'Nam');

-- Technicians at d2 (Ha Noi)
INSERT INTO technicians (id, dealership_id, name) VALUES
('t4', 'd2', 'Duc');
INSERT INTO technicians (id, dealership_id, name) VALUES
('t5', 'd2', 'Hoang');

-- Qualifications — varied per dealership
-- d1: Minh (s1,s2,s4), Hai (s1,s3), Nam (s2,s3)
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t1', 's1'), ('t1', 's2'), ('t1', 's4');
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t2', 's1'), ('t2', 's3');
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t3', 's2'), ('t3', 's3');
-- d2: Duc (s1,s2,s4), Hoang (s1,s3)
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t4', 's1'), ('t4', 's2'), ('t4', 's4');
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t5', 's1'), ('t5', 's3');

-- Service Bays at d1
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b11', 'd1', 'Bay 1A');
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b12', 'd1', 'Bay 2A');
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b13', 'd1', 'Bay 3A');

-- Service Bays at d2
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b21', 'd2', 'Bay 1B');
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b22', 'd2', 'Bay 2B');
