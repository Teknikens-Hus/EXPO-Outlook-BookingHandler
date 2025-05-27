FROM golang:1.24.3 AS build-stage

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -C /app/cmd/EXPO-Outlook-BookingHandler -o /EXPO-Outlook-BookingHandler

# Run the tests in the container
#FROM build-stage AS run-test-stage
#RUN go test -v ./...

# Deploy the application binary into a lean image
FROM alpine:latest AS build-release-stage

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary
COPY --from=build-stage /EXPO-Outlook-BookingHandler ./EXPO-Outlook-BookingHandler
# Copy the query-booking.graphql file
COPY --from=build-stage /app/cmd/EXPO-Outlook-BookingHandler/query-booking.graphql ./query-booking.graphql

RUN chown -R appuser:appgroup /app

USER appuser

ENTRYPOINT ["/app/EXPO-Outlook-BookingHandler"]