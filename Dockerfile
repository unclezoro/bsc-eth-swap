FROM ubuntu:18.04

# Set working directory for the build
WORKDIR /opt/app

# Add source files
COPY ./build/swap-backend /opt/app

# Run as non-root user for security
USER 1000

# Run the app
CMD /opt/app/swap-backend

