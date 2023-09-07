FROM golang:1.18-alpine

ENV PORT=80
ENV env=dev

# Install Xvfb and other dependencies
RUN apk add --no-cache xvfb chromium chromium-chromedriver openjdk11 gcc musl-dev make

ENV JAVA_HOME=/usr/lib/jvm/default-jvm

# Set environment variables for Xvfb and ChromeDriver
ENV DISPLAY=:99
ENV CHROME_DRIVER_PATH="/usr/lib/chromium/chromedriver"

# Set up Xvfb
RUN Xvfb :99 -screen 0 1280x720x24 -ac +extension GLX +render -noreset &

WORKDIR /go/src/twitter-bot

COPY . .

RUN go mod download && go install github.com/cosmtrek/air@latest

CMD ["air", "."]