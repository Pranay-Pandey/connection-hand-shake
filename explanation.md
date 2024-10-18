# Logistics Platform Explanation Document

We use a microservices architecture to build a logistics platform that connects customers with drivers for on-demand transportation services. The platform consists of several services that handle different aspects of the transportation process, such as booking requests, driver tracking and pricing.

We have the following services:

1. **Authentication Service**: Handles user/driver authentication and authorization. It generates and verifies JWT tokens for secure communication between services. This JWT token is used to authenticate the user/driver for all the services.
    *Databases*:
        - PostgreSQL: Stores user/driver/admin credentials.

2. **Booking Service**: Manages the booking process, including creating, updating, and accepting booking requests. It also handles booking request expiration and notifications generation. The booking service is the source of truth for all booking-related data. It accepts user requests, handles driver allocation and updates the booking status.
    *Databases*:
        - MongoDB: Stores booking requests and related data.
        - Redis: Stores active driver pool for quick retrieval.
        - Postgres: Stores booking status and history.
    We do not store the booking requests in the postgres permanent database, only the final accepted bookings are stored in the postgres database.

3. **Driver Location**: This service consumes driver location updates from the notification service and updates the active driver pool in Redis. It does not have any exposed endpoints.
    *Databases*:
        - Redis: Stores active driver pool for quick retrieval.

4. **Notification Service**: Handles real-time communication between users and drivers. It uses WebSockets to provide real-time updates on booking status, driver location, and other notifications.It managers both user and driver connections and sends notifications to both parties. It stores the connection of driver and user as well as the relation of active booking driver-user in memory.

5. **Pricing Service**: Provides cost estimates for transportation based on distance, time, and other factors. It is used by the booking service to calculate the price for a booking request. It also handles price surges during peak times.
    *Databases*:
        - Redis: Stores active driver pool for quick retrieval. This is used to estimate the demand and apply surge pricing accordingly.

6. **Admin Service**: Provides analytics and insights into the platform's performance, including driver and fleet statistics, booking analytics, and vehicle locations. It is used by administrators to monitor and manage the logistics platform. 
    
## Communication Between Services

We utilize the following Kafka topics for communication between services:

1. **driver_notification**: Produced by the booking service to notify the notification service of new booking requests that need to be matched with drivers. Consumed by the notification service to send notifications to drivers. 

2. **driver_locations**: Produced by the notification service to update the driver's location. The notification service receives the driver's location from the active driver and sends it to the driver_location service. Consumed by the driver location service to update the active driver pool in Redis. This message is only produced if the driver is idle and not in an active booking.

3. **booking_notifications**: Produced by the booking service to notify the notification service of accepted booking requests. 
    -   Consumed by the notification service to send notifications to users and drivers. This also establishes the connection between user and driver for real-time updates.
    -   Consumed by the driver location service to remove the driver from the active driver pool. This is done when the driver accepts the booking request. We will again get the update from booking service when the transport is completed and add the driver to the active driver pool.


## Database Schema

1. **PostgreSQL**:
    - **User**: id, name, email, password, created_at, updated_at
    - **Admin**: id, name, email, password, created_at, updated_at
    - **VehicleDriver**: id, name, vehicleId, email, password, vehicleType, vehicleVolume
    - **Booking**: id, userId, driverId, pickupLocation, dropoffLocation, price, status, created_at, completed_at

2. **MongoDB**:
    - **BookingRequest**: userId, userName, pickupLocation, dropoffLocation, price, created_at, vehicleType
    - **DriverLocation**: driverId, location, timestamp -- store the driver location in MongoDB as well for backup and audit purposes, as a feature.

3. **Redis**:
    - **ActiveDriverPool**: driverId, location -- stores the active driver pool for quick retrieval based on location.

## Real-time Communication

We use WebSockets to provide real-time updates to users and drivers. The notification service maintains WebSocket connections with users and drivers to send updates on booking status, driver location, and other notifications. This allows for efficient and low-latency communication between the platform and its users.

## Flow of Events

1. **User Flow**:
    - User makes a booking request through the booking service.
    - The booking service searches for nearby drivers and sends a driver notification to the notification service.
    - The notification service notifies the driver of the booking request.
    - If the driver accepts the request, the booking service updates the booking status and sends a booking notification to the notification service.
    - The notification service establishes a real-time connection between the user and driver for updates.
    - The driver location service updates the active driver pool in Redis with the driver's location.
    - The notification service sends real-time updates to the user on the driver's location and booking status.
    - Once the transport is completed, the booking service updates the booking status and sends a booking notification to the notification service to remove the driver from the active driver pool.

2. **Driver Flow**:
    - Driver sends location updates to the notification service.
    - The notification service sends the driver's location to the driver location service.
    - The driver location service updates the active driver pool in Redis with the driver's location.
    - If there is a nearby booking request, the booking service sends a driver notification to the notification service.
    - The notification service notifies the driver of the booking request.
    - If the driver accepts the request, the booking service updates the booking status and sends a booking notification to the notification service.
    - The notification service establishes a real-time connection between the user and driver for updates.
    - The notification service sends real-time updates to the driver on the user's location and booking status.
    - Once the transport is completed, the booking service updates the booking status and sends a booking notification to the notification service to remove the driver from the active driver pool.

## Major Design Decisions and Trade-offs

1. **Event-Based Microservice Architecture**: We chose this architecture for its scalability and flexibility. It allows us to independently scale and deploy services, making the system more resilient and easier to maintain. The trade-off is increased complexity in system design and potential eventual consistency issues.

2. **Kafka as Message Queue**: Kafka provides high throughput, fault-tolerance, and horizontal scalability, which are crucial for our high-volume traffic. It also allows for event replay and acts as a buffer during traffic spikes. The trade-off is the additional operational complexity of managing a Kafka cluster.

3. **WebSockets for Real-time Updates**: WebSockets provide bi-directional communication, which is more efficient for real-time updates compared to long polling. This is crucial for features like real-time driver tracking. The trade-off is increased server resource usage for maintaining open connections.

4. **Redis for Active Driver Pool**: Redis's geo-spatial indexing capabilities make it ideal for quickly finding nearby drivers. It's also in-memory, providing fast read/write operations. 

5. **MongoDB for Booking Requests**: MongoDB's flexibility allows us to store booking requests with varying structures. Its TTL index feature is used to automatically expire unaccepted booking requests. The trade-off is eventual consistency, which we mitigate by using the booking service as the source of truth.

6. **PostgreSQL with Sharding**: PostgreSQL provides ACID compliance for critical transactional data. Sharding improves read/write performance and allows for better data distribution. We shard the database according to the location. (The drivers in US need not be concerned about the user requests in India). The trade-off is increased complexity in managing and querying across shards.

7. **Separate Pricing Service**: This allows for independent scaling and rate limiting of the pricing functionality. It also provides flexibility to implement complex pricing models without affecting other services. The trade-off is an additional network hop for pricing calculations. Currently we use a constant config for calculating the price. THe config depends on the distance and time taken for the trip depending on each vehicle type. Price surges at peak times are also implemented.

Some considerations - 

1. Either to use websockets or long polling for realtime updates. We chose websockets as it is more efficient and scalable.
2. Use of Kafka as message queue to decouple the services.
3. Either to use Redis or MongoDB for active driver pool. We chose Redis as it is more efficient for this use case.
4. We also have a separate service for pricing as it can be scaled independently and can be used by other services as well. We can apply rate limiting to this service individually to prevent abuse.
5. Add quicker booking request validation we get the mongoDB entry ID and check if it is present in the booking service. ALso we have added an index that will expire the booking request after 10 minutes if it is not accepted by any driver.
6. Even if we have more than 1 driver accepting the request, we use kafka to maintain the order of the requests and notify the user accordingly.
7. For pricing we use the active driver pool to get estimate of demand and apply surge accordingly.



## Handling High-Volume Traffic

1. **Horizontal Scaling**: All services are seperate, allowing us to scale horizontally by adding more instances as needed.

2. **Caching**: We use Redis not only for the active driver pool but also as a cache for frequently accessed data to reduce database load. Example - We map the vehicle type with driver ID

3. **Database Sharding**: PostgreSQL is sharded based on location (city or country level) to distribute the database load.

4. **Asynchronous Processing**: Non-critical operations are processed asynchronously using Kafka, reducing response times and allowing for better handling of traffic spikes.
