# TCP Server with Proof Of Work

This project is a TCP server that implements a Proof of Work (PoW) algorithm to protect against DDoS attacks. Upon successful PoW solution, the client receives a random quote.

## Running the Project

```bash
make start
```

This command starts the TCP server, which listens for connections on port 10001.

Several clients will also be started to test the service. Most of the clients will send valid PoW solutions, but for demonstration purposes, a small percentage of the clients will send mock erroneous solutions.

Logs can be viewed using Docker Desktop.

## Key Components

### Hashcash

[Interface](internal/pow/types.go)

The Proof of Work algorithm used is Hashcash, chosen for its simplicity and frequent use in similar scenarios.

[Implementation](internal/pow/hashcash/hashcash.go)

### Quotes

[Interface](internal/server/quotes/types.go)

This service returns a random quote from a predefined set. For simplicity, the quotes are stored in memory as a slice.

[Implementation](internal/server/quotes/inmemory/inmemory.go)

### Server

A TCP server for handling client connections. Upon a client's connection, the server sends a new challenge and waits for the solution. On success, the server sends one of the quotes. On failure, the connection is closed.

### Client

A client that connects to the server and is capable of solving PoW challenges.

## Configuration

Both the server and client applications are configured using environment variables.

[Server Configuration](internal/server/config/config.go)

[Client Configuration](internal/client/config/config.go)

## Testing

To run unit tests:

```bash
make test
```

Unit tests are written for the [hashcash module](internal/pow/hashcash/).

To run functional tests:

```bash
make func-tests
```

Functional tests check two types of interaction with the service:
* Sending a valid solution.
* Sending an incorrect solution.

The goal of this test assignment was not to cover the entire codebase with tests. Therefore, there is a possibility that some edge cases may not be covered by unit or functional tests.

## Debugging

For debugging purposes, you can enable pprof. This allows you to collect profiles and analyze performance. To enable pprof, set the PPROF_ENABLED environment variable to true.

## Linting

To run the linter:

```bash
make lint
```
