# Specifies a parent image
FROM justb4/jmeter

# Creates an app directory to hold your app’s source code

WORKDIR /

# Copies everything from your root directory into /app
COPY HTTP_Requests.jmx /HTTP_Requests.jmx

# Installs Go dependencies
#RUN go mod download

# Builds your app with optional configuration
#RUN go build -o /godocker

# Tells Docker which network port your container listens on
# EXPOSE 8000

# Specifies the executable command that runs when the container starts
CMD ["-n", "-t", "/HTTP_Requests.jmx","-l","jmeter.jtl","-j","/jmeter/jmeter.log" ]
