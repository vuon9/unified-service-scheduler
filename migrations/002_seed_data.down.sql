DELETE FROM service_bays WHERE dealership_id = 'd1';
DELETE FROM technician_qualifications WHERE technician_id IN ('t1', 't2', 't3');
DELETE FROM technicians WHERE dealership_id = 'd1';
DELETE FROM service_types WHERE id IN ('s1', 's2', 's3');
DELETE FROM vehicles WHERE id IN ('v1', 'v2');
DELETE FROM customers WHERE id IN ('c1', 'c2');
DELETE FROM dealerships WHERE id = 'd1';
