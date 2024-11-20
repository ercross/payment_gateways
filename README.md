# Payment Gateway Integration

## Overview
The **Payment Gateway Integration** project is designed to facilitate seamless deposit and withdrawal operations within a trading system
while providing status callbacks for transaction processing. 
It emphasizes modularity, fault tolerance, and an intuitive API design to streamline payment workflows.


---

## Features

1. **Deposit**
    - Allows users to add funds to their account.
2. **Withdrawal**
    - Handles requests to withdraw funds.
3. **Transaction Status Callback**
    - Updates the status of transactions asynchronously.

---

## Components
The project comprises the following components:
1. **API Service**
    - Written in Go, this service handles payment initiation (deposits/withdrawals) and transaction callbacks.
2. **PostgreSQL Database**
    - Stores user, transaction, and gateway-related data.
3. **Redis Cache**
    - Provides distributed rate limiting, global lock for idempotent transactions, and caching for enhanced performance.
4. **Kafka**
    - Ensures reliable and asynchronous communication for payment-related events.
5. **Zookeeper**
    - Manages Kafka brokers.
6. **Docker Compose**
    - Facilitates containerized deployments using bind mounts for local development.

---

## Non-Functional Requirements
This project aims to achieve the following system design goals:

- **Fault Tolerance:**  
  Ensures minimal downtime through proper retry mechanisms and distributed systems like Kafka.

- **Resiliency:**  
  Incorporates caching, message queues, and database transactions to recover from transient failures.

- **Data Security:**  
  Data is securely managed using industry standards, though authentication and session handling are minimal.

- **Modularity:**  
  A feature-based architecture organizes components that interact closely together, making the system easier to extend and maintain.

---

## Architectural Decisions and Assumptions

1. **Modular, Feature-Based Design:**
    - Components interacting closely (e.g., api and data transfer objects) are grouped together.
    - This approach makes it easier to extend specific features independently.

2. **Predefined Users:**
    - Authentication and session handling are minimal to focus on payment processing.
    - Users interact with the system by assuming identities of predefined users seeded into database at startup.

3. **Bounded Contexts:**
    - Business logic, such as withdrawals and deposits, is isolated to reduce the risk of unintended interference.

4. **Technology Choices:**
    - PostgreSQL is chosen for its reliability and relational data handling.
    - Redis is used for caching and distributed rate limiting to ensure performance under load.
    - Kafka enables event-driven processing for scalability.

---

## Getting Started

Follow these steps to set up the project locally:

### Prerequisites
- Docker and Docker Compose installed.
- Make sure `make` is available on your system.

### Setup Instructions

1. **Clone the Repository:**
   ```bash  
   git clone https://github.com/ercross/payment_gateways.git  
   cd payment-gateways
   ```  

2. **Start the Services:**  
   Use the `Makefile` to deploy the project with Docker Compose:
   ```bash  
   make deploy
   ```  

   This starts all services, including the API, PostgreSQL, Redis, Kafka, and Zookeeper.

3**Access the API:**
    - The API is available at `http://localhost:15001`.
    - Use tools like Postman or cURL to interact with the endpoints.

### Tear Down
To stop and remove all Docker containers, networks, and volumes:
```bash  
make clean
```

---

## Limitations

- Minimal authentication: Users assume identities of seeded users.
- Intended for demonstration purposes; real-world security concerns like token-based authentication are not addressed.
