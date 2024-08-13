FROM ubuntu:24.04

# Update and install required packages
RUN apt-get update && \
    apt-get install -y \
    wget \
    gnupg \
    lsb-release \
    cron \
    sudo \
    tar \
    gzip

# Add PostgreSQL APT repository and key
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
    
# Install Go
RUN wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz

# Set up Go environment
ENV PATH="/usr/local/go/bin:${PATH}"

# Install PostgreSQL client
RUN apt-get update && \
    apt-get install -y postgresql-client-16

# Install Go application
RUN go install github.com/RBASWE/db-backup-runner@latest

ENV LOG_FILE_PATH=/var/log/db-backup-runner.log

# Set the default command
CMD ["cron", "-f"]
