-- Seed appointments — all non-overlapping (verified)
-- Times in UTC (Z) for consistent SQLite string comparison

-- Jul 20 (Monday) — Saigon Auto
INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id, technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
VALUES
('a01', 'c1', 'v1', 'd1', 's1', 't1', 'b1', '2026-07-20T01:00:00Z', '2026-07-20T02:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a02', 'c2', 'v2', 'd1', 's3', 't2', 'b2', '2026-07-20T01:00:00Z', '2026-07-20T02:30:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a03', 'c3', 'v3', 'd1', 's2', 't3', 'b1', '2026-07-20T02:00:00Z', '2026-07-20T04:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a04', 'c1', 'v1', 'd1', 's1', 't2', 'b2', '2026-07-20T02:30:00Z', '2026-07-20T03:30:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a05', 'c2', 'v2', 'd1', 's2', 't1', 'b2', '2026-07-20T03:30:00Z', '2026-07-20T05:30:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a06', 'c3', 'v3', 'd1', 's1', 't2', 'b1', '2026-07-20T04:00:00Z', '2026-07-20T05:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a07', 'c1', 'v1', 'd1', 's1', 't1', 'b1', '2026-07-20T06:00:00Z', '2026-07-20T07:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a08', 'c2', 'v2', 'd1', 's3', 't2', 'b2', '2026-07-20T06:00:00Z', '2026-07-20T07:30:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a09', 'c3', 'v3', 'd1', 's2', 't3', 'b1', '2026-07-20T07:00:00Z', '2026-07-20T09:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a10', 'c1', 'v1', 'd1', 's1', 't1', 'b2', '2026-07-20T07:30:00Z', '2026-07-20T08:30:00Z', 'confirmed', CURRENT_TIMESTAMP);

-- Jul 21 (Tuesday)
INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id, technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
VALUES
('a11', 'c3', 'v3', 'd1', 's1', 't1', 'b1', '2026-07-21T01:00:00Z', '2026-07-21T02:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a12', 'c2', 'v2', 'd1', 's3', 't2', 'b2', '2026-07-21T01:00:00Z', '2026-07-21T02:30:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a13', 'c1', 'v1', 'd1', 's2', 't3', 'b1', '2026-07-21T02:00:00Z', '2026-07-21T04:00:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a14', 'c3', 'v3', 'd1', 's1', 't2', 'b2', '2026-07-21T02:30:00Z', '2026-07-21T03:30:00Z', 'confirmed', CURRENT_TIMESTAMP),
('a15', 'c1', 'v1', 'd1', 's1', 't1', 'b1', '2026-07-21T06:00:00Z', '2026-07-21T07:00:00Z', 'confirmed', CURRENT_TIMESTAMP);

-- One cancelled appointment for demo
INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id, technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
VALUES
('a16', 'c2', 'v2', 'd1', 's1', 't1', 'b1', '2026-07-22T02:00:00Z', '2026-07-22T03:00:00Z', 'cancelled', CURRENT_TIMESTAMP);
