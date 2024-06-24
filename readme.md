# SchedulerV2 Service

## Overview
SchedulerV2 is a robust scheduling service designed to manage and process scheduled tasks efficiently. It leverages a microservice architecture to ensure scalability and reliability. The service is built with Go and utilizes a variety of libraries and tools to maintain high performance and ease of use.

## Features
- **Task Scheduling:** Ability to schedule tasks with a specified retry logic.
- **ZooKeeper Integration:** Utilizes Apache ZooKeeper for distributed coordination and configuration management.
- **Database Support:** Integrates with PostgreSQL using GORM for object-relational mapping.
- **Environment Configuration:** Supports environment-based configuration for easy deployment across different stages.
- **Docker Support:** Containerized with Docker for simplified scaling and deployment.
- **Callback Mechanism:** Provides a callback system to notify client services upon task completion or failure.

## Prerequisites
Before you begin, ensure you have met the following requirements:
- Go version 1.20 or above is installed.
- Docker is installed if you wish to run the service in a container.
- Access to a PostgreSQL database.
- An Apache ZooKeeper cluster is set up if you wish to use distributed coordination features.

## Installation
To install SchedulerV2, follow these steps:
1. Clone the repository to your local machine.
2. Navigate to the project directory.
3. Compile the service using the Go compiler with the command `go build`.
4. Set the necessary environment variables such as `database_dsn`, `zookeeper_host`, `dlq_message_limit` and `messages_limit`.

## Configuration
The service can be configured using environment variables. Key configurations include:
- `database_dsn`: The data source name for connecting to the PostgreSQL database.
- `zookeeper_hosts`: Comma-separated list of ZooKeeper hosts.
- `messages_limit`: The no of `PENDING` messages to be processed each second
- `dlq_message_limit`: This the count after which you want your messages to be moved to the dlq table to avoid unlimited retry.
- `zookeeper_heart_beat_time`: This is the session time for the zookeeper session.

## Running the Service
To run SchedulerV2, execute the compiled binary with the command `./schedulerV2`.

Alternatively, you can use Docker to build and run the service with the command `docker-compose up`.

## API Reference
SchedulerV2 exposes a RESTful API for interacting with the service. The API documentation is provided separately.

## Development
For development purposes, you can use the provided Dockerfile to build a local image of the service. Refer to the Dockerfile for details on the build process.

## Contributing
Contributions to the SchedulerV2 project are welcome. Please adhere to the coding conventions and commit guidelines.

## License
SchedulerV2 is released under the MIT License. See the LICENSE file for more details.

## Contact
If you have any questions or feedback, please reach out to the project maintainers.

---
This README provides a basic introduction to the SchedulerV2 service. For more detailed information, please refer to the source code and additional documentation provided within the project repository.
