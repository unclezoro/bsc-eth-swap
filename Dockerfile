FROM golang:1.13-alpine

# Set up apk dependencies
ENV PACKAGES make git libc-dev bash gcc linux-headers eudev-dev curl ca-certificates

COPY .netrc /root/.netrc
RUN chmod 600 /root/.netrc

# Set working directory for the build
WORKDIR /opt/app

# Add source files
COPY . .

# Install minimum necessary dependencies, remove packages
RUN apk add --no-cache $PACKAGES && \
    make build

# Run as non-root user for security
USER 1000

# Run the app
CMD ./build/swap-backend --config-type aws  --aws-region ${AWS_REGION}  --aws-secret-key  ${AWS_SECRET_KEY}

