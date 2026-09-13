FROM golang:latest

# Create the 'dev' user with a home directory
RUN useradd -m -s /bin/bash dev && \
    # Ensure the go module and build cache directories exist with proper permissions
    mkdir -p /home/dev/go/pkg/mod /home/dev/.cache/go-build && \
    chown -R dev:dev /home/dev

RUN cd /usr/local/bin && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xz && \
    chmod +x migrate

# Switch to dev user (optional, but keeps the security model consistent)
USER dev