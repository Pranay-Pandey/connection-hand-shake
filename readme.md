# Connection-Hand-Shake

A logistic platform build to scale.
A microservice architecture that is scalable and fault tolerant. 
For more information on the project, please refer to the [explanation](https://github.com/Pranay-Pandey/connection-hand-shake/blob/main/explanation.md#logistics-platform-explanation-document)


To run it locally, 
you must have a mongoDB connection string - either locally or cloud based
you must have a redis connection string - either locally or cloud based
Kafka running on some port

1. Clone the repository
2. Fill the .env file with the required credentials for the database and kafka
3. Run the docker-compose command to start the services 
```bash
docker-compose build
```
4. Run the docker-compose command to start the services for accessing the database
```bash
docker-compose up -d
```

5. Make migrations
```bash
sh migration.sh
```

6. Run the services
```bash
docker-compose up
```

The services will be up and running on the specified ports
