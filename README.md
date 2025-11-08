# ThreeDoors Application

[![Go Report Card](https://goreportcard.com/badge/github.com/arcaven/ThreeDoors)](https://goreportcard.com/report/github.com/arcaven/ThreeDoors)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://example.com/your-build-status)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)](https://example.com/your-test-status)

## Introduction

ThreeDoors is a minimalist task management application designed to help users organize their daily tasks efficiently. Built with a focus on simplicity and speed, it provides a straightforward interface for adding, completing, and tracking your to-do items. This project serves as a foundational example for building robust, scalable applications using modern development practices.

## Project Status

We are currently in the early stages of development, focusing on core functionalities.
**Current Milestone:** Epic 1, Story 1.2 - Implementing basic task creation and listing.
**Next Goal:** Working towards a tech demo showcasing fundamental CRUD operations for tasks.

## Getting Started

To get a local copy up and running, follow these simple steps.

### Prerequisites

Ensure you have the following installed:
* Go (version 1.20 or higher)
* Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/arcaven/ThreeDoors.git
   cd simple-todo
   ```
2. Build the application:
   ```bash
   go build -o threedoors ./cmd/threedoors
   ```
3. Run the application:
   ```bash
   ./threedoors
   ```

## Setting up for Development

### Dependencies

This project uses Go modules for dependency management. Dependencies are automatically downloaded when you build or run the project.

### Running Tests

To run the unit tests:
```bash
go test ./...
```

### Code Style and Linting

We adhere to standard Go formatting and linting practices. Please ensure your code is formatted and passes lint checks before submitting a pull request.
```bash
go fmt ./...
go vet ./...
```

## Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please ensure your pull requests adhere to our coding standards and include appropriate tests.

## Documentation

For more detailed information about the project's architecture, design decisions, and product requirements, please refer to our comprehensive documentation:

*   **Architecture Documentation:** [docs/architecture/index.md](docs/architecture/index.md)
*   **Product Requirements Document (PRD):** [docs/prd/index.md](docs/prd/index.md)

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Your Name/Team Name - your_email@example.com
Project Link: [https://github.com/arcaven/ThreeDoors](https://github.com/arcaven/ThreeDoors)
