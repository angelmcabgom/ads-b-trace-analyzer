FROM golang:latest

# Create the 'dev' user with a home directory
RUN useradd -m -s /bin/bash dev && \
    # Ensure the go module and build cache directories exist with proper permissions
    mkdir -p /home/dev/go/pkg/mod /home/dev/.cache/go-build && \
    chown -R dev:dev /home/dev

RUN cd /usr/local/bin && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xz && \
    chmod +x migrate

RUN apt update && apt install -y postgresql && apt install -y iputils-ping

# RUN go install -v github.com/go-delve/delve/cmd/dlv@latest

# Switch to dev user (optional, but keeps the security model consistent)
USER dev