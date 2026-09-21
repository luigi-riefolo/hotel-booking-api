# Hotel Booking API

A small REST API for booking hotel rooms, written in Go with Echo, GORM and
PostgreSQL.

## How to run it

Requires docker-compose and newman (for testing).

```
make up      # start the api on :8080 and postgres on :5432
make down    # stop and remove the data
```

To run the API outside Docker, start just the database with
`docker compose up db` and then `make run`.

Configuration is read from the environment: `PORT` (default `8080`),
`LOG_LEVEL` (`info`) and `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`,
`POSTGRES_PASSWORD`, `POSTGRES_DB_NAME` (defaults match the compose file).
See `./internal/config/config.go` for more details


## API docs

You can find the api docs [here](./docs/openapi.yaml)

## Endpoints

| Method | Path                                   | Description                                 |
|--------|----------------------------------------|---------------------------------------------|
| GET    | `/api/v1/hotels?name=`                 | Find hotels by (partial) name, with rooms   |
| GET    | `/api/v1/hotels/{id}/rooms?check_in=&check_out=&guests=` | Rooms free for the whole stay |
| POST   | `/api/v1/bookings`                     | Book a room                                 |
| GET    | `/api/v1/bookings/{reference}`         | Get a booking                               |
| POST   | `/api/v1/test/seed`                    | Reset and load three hotels with six rooms  |
| POST   | `/api/v1/test/reset`                   | Remove all data                             |

Dates are `YYYY-MM-DD`. A stay covers the nights from `check_in` up to, but not
including, `check_out`, so a room can be booked from the day someone else
checks out.


Example:

```
curl -X POST localhost:8080/api/v1/test/seed

curl "localhost:8080/api/v1/hotels?name=overlook"

curl "localhost:8080/api/v1/hotels/1/rooms?check_in=2026-12-01&check_out=2026-12-04&guests=2"

curl -X POST localhost:8080/api/v1/bookings -H 'Content-Type: application/json' -d '{
  "room_id": 3, "guests": 2,
  "guest_name": "Homer Simpson", "guest_email": "homer.simpson@springfield.com",
  "check_in": "2026-12-01", "check_out": "2026-12-04"
}'

curl localhost:8080/api/v1/bookings/BK-XXXXXXXXXXXXXXXX
```

Errors come back as `{"message": "..."}` with `400` for bad input, `404` for
an unknown room or booking, `409` if the room is already taken and `422` if
the party is too big for the room.

## How the rules are enforced

- **Six rooms, three types**: the seed gives each hotel two single (sleeps 1),
  two double (2) and two deluxe (4) rooms.

- **No double booking**: booking runs in a transaction that locks the room row
  (`SELECT ... FOR UPDATE`) before checking for overlapping bookings, so
  concurrent requests for the same room are serialised

- **No changing rooms**: a booking is always for one room, and availability
  only returns rooms that are free for every night requested.

- **Capacity**: the party size is checked against the room's capacity

- **Unique references**: random `BK-` references backed by a unique index.

- **Confirmation email**: after a booking a goroutine waits two seconds and
  logs `confirmation email sent`. On shutdown the server waits for pending
  emails before exiting.



## Testing

```
make test     # unit tests
make newman   # Postman collection in tests/, against a running api
```

The Postman collection can also be imported into Postman directly. It seeds
the database, walks through the booking flow and checks each rule.

## TODO

- Cancelling bookings
- Pricing
- Integration tests against a real database for the booking transaction
- Pagination on hotel search


# hotel-booking-api
# hotel-booking-api
