-- Dealership
INSERT INTO dealerships (id, name, address, opening_hours) VALUES
('d1', 'Saigon Auto', '123 Nguyen Hue, District 1, Ho Chi Minh City', '{"mon":"08:00-17:00","tue":"08:00-17:00","wed":"08:00-17:00","thu":"08:00-17:00","fri":"08:00-17:00","sat":"08:00-12:00","sun":"closed"}');

-- Customers
INSERT INTO customers (id, name, email, phone) VALUES
('c1', 'Anh Tuan', 'anhtuan@example.com', '0901234567');
INSERT INTO customers (id, name, email, phone) VALUES
('c2', 'Chi Lan', 'chilan@example.com', '0907654321');

-- Vehicles
INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES
('v1', 'c1', 'VIN-TOYOTA-CAMRY-2023', 'Toyota', 'Camry', 2023);
INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES
('v2', 'c2', 'VIN-HONDA-CIVIC-2022', 'Honda', 'Civic', 2022);

-- Service Types
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s1', 'Oil Change', 60, 'Full synthetic oil change with filter replacement');
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s2', 'Brake Replacement', 120, 'Front and rear brake pad replacement');
INSERT INTO service_types (id, name, duration_minutes, description) VALUES
('s3', 'Engine Diagnostic', 90, 'Comprehensive engine diagnostic scan and report');

-- Technicians at d1
INSERT INTO technicians (id, dealership_id, name) VALUES
('t1', 'd1', 'Minh');
INSERT INTO technicians (id, dealership_id, name) VALUES
('t2', 'd1', 'Hai');
INSERT INTO technicians (id, dealership_id, name) VALUES
('t3', 'd1', 'Nam');

-- Technician Qualifications
-- t1 (Minh): qualified for s1 (Oil Change) and s2 (Brake Replacement)
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t1', 's1');
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t1', 's2');
-- t2 (Hai): qualified for s1 (Oil Change) and s3 (Engine Diagnostic)
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t2', 's1');
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t2', 's3');
-- t3 (Nam): qualified for s2 (Brake Replacement)
INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES
('t3', 's2');

-- Customer 3 (for 3-car demo)
INSERT INTO customers (id, name, email, phone) VALUES
('c3', 'Bao Minh', 'baominh@example.com', '0901122334');

-- Vehicle 3
INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES
('v3', 'c3', 'VIN-MAZDA-CX5-2024', 'Mazda', 'CX-5', 2024);

-- Service Bays at d1
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b1', 'd1', 'Bay 1');
INSERT INTO service_bays (id, dealership_id, name) VALUES
('b2', 'd1', 'Bay 2');
