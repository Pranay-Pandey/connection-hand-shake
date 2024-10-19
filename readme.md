# Connection-Hand-Shake

A logistic platform build to scale.
A microservice architecture that is scalable and fault tolerant. 
For more information on the project, please refer to the [explanation](https://github.com/Pranay-Pandey/connection-hand-shake/blob/main/explanation.md#logistics-platform-explanation-document)


To run it locally, you should have

* MongoDB connection string - either locally or cloud based. To be provided in the .env file
* Redis connection string - either locally or cloud based. To be provided in the .env file

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

6. Check the services
```bash
docker-compose up
```

The services will be up and running on the specified ports


# 

To run the frontend of the project. Follow the steps

1. Go to the frontend directory
```bash
cd frontend
```

2. Install the dependencies
```bash
npm install
```

3. Run the frontend
```bash
npm run dev
```

The frontend will be up and running on the specified port