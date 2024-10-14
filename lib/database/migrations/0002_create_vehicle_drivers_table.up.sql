CREATE TABLE IF NOT EXISTS vehicle_drivers (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    vehicleID     VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password      VARCHAR(255) NOT NULL,
    vehicleType   VARCHAR(255) NOT NULL,
    vehicleVolume VARCHAR(255) NOT NULL
);