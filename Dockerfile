FROM golang:1.13-alpine

# Set up apk dependencies
ENV PACKAGES make git libc-dev bash gcc linux-headers eudev-dev curl ca-certificates
# Set working directory for the build
WORKDIR /opt/app

# Add source files
COPY ./build/swap-backend /opt/app

# Run as non-root user for security
USER 1000

# Run the app
CMD /opt/app/swap-backend --config-type aws  --aws-region ${AWS_REGION}  --aws-secret-key  ${AWS_SECRET_KEY}

