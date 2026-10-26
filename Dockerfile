FROM node:25 AS libs

WORKDIR /build

COPY package.json package-lock.json ./

RUN npm ci --ignore-scripts \
    && npm run build:libs

# ---

FROM golang:1.27.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./
COPY videos videos
COPY static static
COPY --from=libs /build/static/lib static/lib

RUN CGO_ENABLED=0 go build -trimpath -o /irmatube .

# ---

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /irmatube /irmatube

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/irmatube"]
