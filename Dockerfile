FROM golang:1.22

# Update and install required packages
RUN apt-get update && \
    apt-get install -y \
    wget \
    gnupg \
    lsb-release \
    cron \
    sudo

# Add PostgreSQL APT repository and key
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -

# Install PostgreSQL client
RUN apt-get update && \
    apt-get install -y postgresql-client-16

# Install Go application
RUN go install github.com/RBASWE/db-backup-runner@0.0.8

# ENV LOG_FILE_PATH=/var/log/db-backup-runner.log

# Set the default command
CMD ["cron", "-f"]
