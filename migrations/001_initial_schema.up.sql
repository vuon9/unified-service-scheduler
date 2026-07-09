CREATE TABLE customers (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT
);

CREATE TABLE vehicles (
    id          TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(id),
    vin         TEXT,
    make        TEXT NOT NULL,
    model       TEXT NOT NULL,
    year        INTEGER NOT NULL
);

CREATE TABLE dealerships (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    address       TEXT NOT NULL,
    opening_hours TEXT
);

CREATE TABLE service_types (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    description      TEXT
);

CREATE TABLE technicians (
    id            TEXT PRIMARY KEY,
    dealership_id TEXT NOT NULL REFERENCES dealerships(id),
    name          TEXT NOT NULL
);

CREATE TABLE technician_qualifications (
    technician_id   TEXT NOT NULL REFERENCES technicians(id),
    service_type_id TEXT NOT NULL REFERENCES service_types(id),
    PRIMARY KEY (technician_id, service_type_id)
);

CREATE TABLE service_bays (
    id            TEXT PRIMARY KEY,
    dealership_id TEXT NOT NULL REFERENCES dealerships(id),
    name          TEXT NOT NULL
);

CREATE TABLE appointments (
    id              TEXT PRIMARY KEY,
    customer_id     TEXT NOT NULL REFERENCES customers(id),
    vehicle_id      TEXT NOT NULL REFERENCES vehicles(id),
    dealership_id   TEXT NOT NULL REFERENCES dealerships(id),
    service_type_id TEXT NOT NULL REFERENCES service_types(id),
    technician_id   TEXT NOT NULL REFERENCES technicians(id),
    service_bay_id  TEXT NOT NULL REFERENCES service_bays(id),
    scheduled_start DATETIME NOT NULL,
    scheduled_end   DATETIME NOT NULL,
    status          TEXT NOT NULL DEFAULT 'confirmed'
                    CHECK(status IN ('confirmed','cancelled','completed')),
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_appointments_technician_time
    ON appointments(technician_id, status, scheduled_start, scheduled_end);

CREATE INDEX idx_appointments_bay_time
    ON appointments(service_bay_id, status, scheduled_start, scheduled_end);
