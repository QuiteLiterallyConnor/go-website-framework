FROM ubuntu:latest
FROM golang:1.20

WORKDIR /resume_website

# Copy application files
COPY main main
COPY static static
COPY templates templates

EXPOSE 8080

# Start the application and LocalXpose with a reserved domain
CMD ["sh", "-c", "./main -port 8080"]

