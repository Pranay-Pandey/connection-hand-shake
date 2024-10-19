#!/bin/bash

# Variables
MASTER="master"
WORKER1="worker1"
WORKER2="worker2"
DB_USER="user"
DB_NAME="citus"

# Add worker nodes
docker-compose exec $MASTER psql -U $DB_USER -d $DB_NAME -c "SELECT master_add_node('$WORKER1', 5432);"
docker-compose exec $MASTER psql -U $DB_USER -d $DB_NAME -c "SELECT master_add_node('$WORKER2', 5432);"

# Create table
docker-compose exec $MASTER psql -U $DB_USER -d $DB_NAME -c "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY,   name VARCHAR(255) NOT NULL,   email VARCHAR(255) UNIQUE NOT NULL,   password VARCHAR(255) NOT NULL,   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP);"
docker-compose exec $MASTER psql -U $DB_USER -d $DB_NAME -c "CREATE TABLE IF NOT EXISTS vehicle_drivers (id SERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, vehicle_id VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, password VARCHAR(255) NOT NULL, vehicle_type VARCHAR(255) NOT NULL, vehicle_volume VARCHAR(255) NOT NULL);"
docker-compose exec $MASTER psql -U $DB_USER -d $DB_NAME -c "CREATE TABLE IF NOT EXISTS booking(id serial, user_id INTEGER NOT NULL,driver_id INTEGER NOT NULL,pickup_latitude FLOAT NOT NULL,pickup_longitude FLOAT NOT NULL,pickup_name VARCHAR(255) NOT NULL,dropoff_latitude FLOAT NOT NULL,dropoff_longitude FLOAT NOT NULL,dropoff_name VARCHAR(255) NOT NULL,vehicle_type VARCHAR(255) NOT NULL,price FLOAT NOT NULL,created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,completed_at TIMESTAMP WITH TIME ZONE,updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,status VARCHAR(255) NOT NULL, PRIMARY KEY(id, pickup_latitude));"


# Distributed table
docker-compose exec $MASTER psql -U $DB_USER -d $DB_NAME -c "SELECT create_distributed_table('booking', 'pickup_latitude');"

echo "Migration complete"