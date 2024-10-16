CREATE TABLE IF NOT EXISTS booking (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    driver_id INTEGER NOT NULL,
    pickup_latitude FLOAT NOT NULL,
    pickup_longitude FLOAT NOT NULL,
    pickup_name VARCHAR(255) NOT NULL,
    dropoff_latitude FLOAT NOT NULL,
    dropoff_longitude FLOAT NOT NULL,
    dropoff_name VARCHAR(255) NOT NULL,
    vehicle_type VARCHAR(255) NOT NULL,
    price FLOAT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(255) NOT NULL
);