# Specifies a parent image
FROM grafana/k6

# Creates an app directory to hold your app’s source code

WORKDIR /

# Copies everything from your root directory into /app
COPY test.js /test.js

# Installs Go dependencies
#RUN go mod download

# Builds your app with optional configuration
#RUN go build -o /godocker

# Tells Docker which network port your container listens on
# EXPOSE 8000

# Specifies the executable command that runs when the container starts
CMD [ "run","/test.js" ]
