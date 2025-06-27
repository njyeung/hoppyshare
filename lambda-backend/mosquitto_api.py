import requests
import os
import subprocess
import json
import paho.mqtt.client as mqtt
import ssl
import time
from utils import api_response

MOSQUITTO_API = "https://localhost"
CERT = "./certs/lambda.crt"
KEY = "./certs/lambda.key"
CA = "./certs/ca.crt"

@api_response
def onboard_user(uid):
    return requests.post(
        f"{MOSQUITTO_API}/onboard_user",
        json={"cn": uid},
        cert=(CERT, KEY),
        verify=CA
    )

@api_response
def add_device(uid):
    return requests.post(
        f"{MOSQUITTO_API}/add_device",
        json={"cn": uid},
        cert=(CERT, KEY),
        verify=CA
    )

@api_response
def revoke_device(cert_pem):
    return requests.post(
        f"{MOSQUITTO_API}/revoke_device",
        json={"cert": cert_pem},
        cert=(CERT, KEY),
        verify=CA
    )

@api_response
def pub_settings(settings: dict, uid: str):
    topic = f"users/{uid}/settings"
    payload = json.dumps(settings)

    client = mqtt.Client()

    try:
        client.tls_set(
            ca_certs=CA,
            certfile=CERT,
            keyfile=KEY,
            tls_version=ssl.PROTOCOL_TLS_CLIENT,
        )
        client.connect("localhost", 8883)
        client.loop_start()

        # Retain message
        result = client.publish(topic, payload, retain=True)
        result.wait_for_publish()

        client.loop_stop()
        client.disconnect()

        return {
            "status_code": 200,
            "json": {"message": "Published", "topic": topic, "payload": settings}
        }

    except Exception as e:
        return {
            "status_code": 500,
            "json": {"error": str(e)}
        }

