FROM eclipse-mosquitto:2.0

COPY mosquitto.conf /mosquitto/config/mosquitto.conf
COPY conf.d /mosquitto/config/conf.d
COPY certs /mosquitto/certs
COPY data /mosquitto/data

EXPOSE 1883 8883

RUN apk update
RUN apk add nginx openssl

COPY nginx/nginx.conf /etc/nginx/nginx.conf

# Add API dependencies
RUN apk add python3 py3-pip

# Copy and install API
COPY api /api
RUN pip3 install --break-system-packages flask gunicorn

USER root

# Run Mosquitto + Flask API
CMD sh -c "mosquitto -c /mosquitto/config/mosquitto.conf & gunicorn --bind 0.0.0.0:5000 api.app:app & nginx -g 'daemon off;'"
