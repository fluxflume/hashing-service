# Hashing Service

## Summary

This project contains a service that accepts a form-encoded "password" field via a POST request, then hashes and
asynchronously persists the value for retrieval at a later time.

## API

### POST /hash and GET /hash/{id}

To hash a password: `curl -d "password=abc123" -X POST localhost:3000/hash`

The response will contain an Id (number) which be can be used to retrieve the value: `curl localhost:3000/hash/0`

> Requests that contain a password longer than 64 characters will fail. See: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#implement-proper-password-strength-controls

### GET /stats

Stats can be retrieved like: `curl localhost:3000/stats`

### /shutdown

A request made to /shutdown will terminate the http server and prevent new requests from being processed. The process
itself will terminate after all pending items are processed.

## (Potentially) Configurable Options

The `src/app/main.go` file contains the main function, and also clearly lays out params that _could_ be exposed as
configurable options:

| Param                  | Description |
| ---------------------- | ------------- |
| `maxPendingItems`      | Maximum number of pending items (hashed, but not yet stored) items allowed at a time. Requests that cause the number of pending items to be beyond this limit will fail. |
| `hashCompletionDelay`  | The delay time before making the hashed value available via `GET /hash/{id}`. |
| `numWorkers`           | The number of goroutines consuming from ethe "hashed items to store" channel. |
| `port`                 | The port that the http server binds to. |

## Requirements

Go 1.17

## Starting the Service

This project uses go modules, which I was unable to run without the strange path argument below:

`go run ./...src/app`

The `./...` "workaround" was referenced in this PR: https://github.com/golang/go/issues/27122

To build the app's executable:

`go build ./...src/app`

Then run: `./app`

## Tests

There are a only few test cases, lots of room for improvement here. The "stubs" that were created in the
model/hashing_test.go file work fine, and the stubs are doing what I hoped, but it's pretty verbose! Refactoring or
cleaning up by using a well known mocking library might help.

## Design Notes

Passwords are hashed immediately to limit the potential for sensitive data leakage. Once hashed (`HashedItem`), items
are "published" (`src/services/pubsub.go`) to be persisted (`HashedItemStore`) so that
`GET /hash/{id}` requests can eventually retrieve and return hashed values.

The design of this service stemmed from the idea of making this service scalable. A few abstractions were created to
support this:

**HashedItemStore**

This interface represents the storage, or database layer. An in-memory implementation exists, `LocalHashedItemStore`. To
support a scalable service though, a new implementation could be created, such as one that's backed by MongoDb or
Firestore.

**IdGenerator**

This interface simply provides the stateful-like, incrementing Ids used in the POST and GET endpoints. If there's only
ever 1 instance of this service, then the provided implementation, `LocalIdGenerator` would suffice. However, if there's
truly a hard requirement on incrementing Ids and this service needed to scale, the source of the Id would need to be
distributed so that the increment operation is atomic. For example, a `RedisIdGenerator` could be implemented to support
scaling beyond 1 instance and provide a consistent, incrementing Id source.

**"PubSub"**

This is probably most "complex" aspect of the service when it comes to scaling, but the provided implementation is
simple.

I thought of the 5 second delay as a side effect of a long-ish running async operation. Going back to the idea of
scaling this service... since the passwords are hashed immediately, the only other work required is the persistence of
the value. This would be required so that all service instances would eventually be able to read the same value for a
given id.

The service receiving the request _could_ just sleep for 5 seconds, wake up and then store the item, but what if during
that time, the database connection became flaky, and the call to store the item failed? Retries could be applied, but if
lots of requests are coming in, those requests will build up and things could get messy.

If instead, a hashed password is immediately published to a distributed messaging system, the message can be consumed by
any service instance, and if there's a failure, the messaging system can re-deliver. It would redeliver until the
message was processed successfully, or worst case, place the message in a dead letter topic for manual intervention.

The delay in this implementation represents a PubSub operations (produce/consume) and the time it takes to store the
item.

The provided implementations make use of goroutines and channels, `LocalPublisher` and `LocalSubscriber`.

**Stats**

This interface is simple, and provides a clean way to update the total count and average. The provided
implementation, `LocalStats` uses a `RWMutex` to ensure the value is safely readable/writable from multiple goroutines.

To support scaling beyond 1 instance of this service, the stats data would be distributed (database, distributed cache
etc.), and an aggregated view of all service instances would be needed.

## Data / Request Flow

### POST / Create Requests

1. Validate provided password and return response error if invalid, otherwise...
2. Generate globally unique id
3. Create `HashedItem`, publish for async processing (delayed by 5 seconds, described below)
4. Return id as response body

### Async Processing

1. Subscriber consumes `HashedItem` and waits 5 seconds
2. Subscriber wakes up and stores the item (in-memory "database")
3. Repeat...

### GET / Retrieve Requests

1. Validate incoming item id and return response error if invalid, otherwise...
2. Fetch the associated `HashedItem` from the store - if it doesn't exist, return a response error, otherwise...
3. Return hashed value as response body

### Stats

Stats are populated as the POST requests are being populated.

### Shutdown

The shutdown process works by publishing a termination signal to a special channel. This channel is read from by the
main goroutine, waiting for a message. Once received, it tells the web server to shutdown and the program exits.

---

## Random Note Collection

I initially thought that I'd use the 5 second delay to simulate processing, mainly the actual string hashing of the
password. If this service was a real production level service, then I could envision publishing the requests to a
messaging system, having consumers pull off messages and perform the hashing, then eventually persist to a database.

But passing around the plaintext password in the app seemed dangerous (it's easy to accidentally log for example), and
even worse would be sending this plaintext password to external systems, across the network.

So I opted for hashing immediately, then using the 5 second delay as a way to simulate an async / eventually consistent
system; making it available to GET requests after the delay.

### Super Simple Solution

Simplest possible solution: immediately hash the incoming value and store it with the creation time. When a request is
made for the item, don't return it unless 5 seconds have passed since creation time!

### Project Structure

I read a few articles on how to structure Go projects. There seems to be somewhat of a standard, but one popular opinion
was to start simple and iterate as needed. I took the latter approach (simple/flat).

I liked the ideas presented in this article, mainly, keeping things simple to begin with (flat)

- https://tutorialedge.net/golang/go-project-structure-best-practices/

I initially wanted the "interfaces" defined separately from implementations,
but having them in the same file turned out to be pretty convenient for this project.

### Challenges

I had no experience with Go before this, except for reading a few articles about it. Most of my time was spent on
learning the language.

The syntax for defining types and implementing interfaces took some time to grok!

Understanding the basics of Go modules took some time.

Still looking into the strange `go build ./...src/app` issue with the hopes of resolving.
