# Builder: golang on Debian trixie, matching go.mod's `go 1.25.0` and the
# rest of this stack's Debian/trixie image family (postgres:17.10-trixie).
FROM golang:1.25.14-trixie AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# pgx is pure Go (no libpq/cgo dependency), so a static binary is enough —
# no C runtime needed in the final image.
RUN CGO_ENABLED=0 GOOS=linux go build -o /agenda-service .

# Runtime: debian:trixie-slim, not distroless — keeps a shell available for
# `docker exec -it` debugging, the same way this project already works
# with the `database` container, at the cost of a slightly larger image
# and a bit more attack surface than a shell-less base. Acceptable here:
# this service is only reachable via the WireGuard tunnel or the trusted
# home LAN, same trust model as the rest of the stack (see CLAUDE.md).
FROM debian:trixie-slim

RUN groupadd --system agenda && useradd --system --gid agenda --no-create-home agenda

COPY --from=builder /agenda-service /usr/local/bin/agenda-service

USER agenda

ENTRYPOINT ["/usr/local/bin/agenda-service"]
