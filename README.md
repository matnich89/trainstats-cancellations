# TrainStats Cancellations

## Overview
TrainStats Cancellations is a service designed to track and manage train cancellations, providing real-time updates and historical data analysis. This service integrates with national rail APIs and manages data using a Redis queue and a relational database.

## Features
- Real-time tracking of train cancellations.
- Historical data analysis of cancellations by date, train ID, and operator.
- Integration with national rail services for up-to-date information.
- Efficient handling of data with Redis and SQL databases.

## Architecture
The application is structured around several key components:
- **HTTP Handlers**: Manage incoming HTTP requests and route them to appropriate controllers.
- **Redis Client**: Handles queue management for incoming train data.
- **National Rail Client**: Integrates with external national rail API.
- **Database**: Stores and retrieves cancellation data.
- **Workers**: Manage background tasks for processing data from the Redis queue.

## Setup
To set up the project locally, follow these steps:
1. Clone the repository.
2. Install dependencies with `go mod tidy`.
3. Set up environment variables or a `.env` file with necessary configurations (Redis, database, API keys).
4. Run the Docker Compose file with `docker-compose up` in the root of the project 
5. Run the migrations to set up the database schema.
6. Start the server with `go run main.go`.

## Usage
Once the application is running, it will start processing incoming data from the configured sources and populate the database with cancellation events.

## License
This project is licensed under the MIT License - see the LICENSE file for details.