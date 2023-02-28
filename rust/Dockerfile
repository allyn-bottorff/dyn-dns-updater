FROM rust:1.67-slim-bullseye as builder
run apt-get update && apt-get install pkg-config libssl-dev -y
RUN mkdir /secrets
WORKDIR /build
COPY ./Cargo.toml .
COPY ./Cargo.lock .
COPY ./src ./src
RUN cargo build --release
COPY ./config.json /secrets/config.json

FROM  debian:bullseye-slim
LABEL org.opencontainers.image.source=https://github.com/allyn-bottorff/dyn-dns-updater
RUN apt-get update && apt-get install ca-certificates -y
RUN useradd -s /bin/nologin --system app
RUN mkdir /secrets
WORKDIR /app
USER app
COPY --from=builder /build/target/release/dyn-dns-updater .
CMD ["/app/dyn-dns-updater"]
