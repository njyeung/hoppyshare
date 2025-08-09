import requests
import os
import subprocess
import json
import paho.mqtt.client as mqtt
import ssl
import time
import boto3
from utils import api_response

MOSQUITTO_API = "https://18.188.110.246"

def download_certs():
    cert_files = {
        "ca.crt": "/tmp/ca.crt",
        "lambda.crt": "/tmp/lambda.crt", 
        "lambda.key": "/tmp/lambda.key"
    }
    
    s3 = boto3.client("s3")
    bucket = "hoppyshare-binaries"
    
    for s3_key, local_path in cert_files.items():
        if not os.path.exists(local_path):
            try:
                s3.download_file(bucket, s3_key, local_path)
                print(f"Downloaded {s3_key} to {local_path}")
            except Exception as e:
                print(f"Failed to download {s3_key}: {e}")
                raise
    
    return "/tmp/lambda.crt", "/tmp/lambda.key", "/tmp/ca.crt"

# Download certificates on module import
try:
    CERT, KEY, CA = download_certs()
# Fallback for local dev
except Exception as e:
    print(f"Failed to download certificates: {e}")
    CERT = "./certs/lambda.crt"
    KEY = "./certs/lambda.key" 
    CA = "./certs/ca.crt"

@api_response
def reload_mosquitto():
    return requests.post(
        f"{MOSQUITTO_API}/reload",
        cert=(CERT, KEY),
        verify=CA
    )

@api_response
def onboard_user(uid):
    return requests.post(
        f"{MOSQUITTO_API}/onboard_user",
        json={"cn": uid},
        cert=(CERT, KEY),
        verify=CA
    )

@api_response
def delete_user(uid):
    return requests.post(
        f"{MOSQUITTO_API}/delete_user",
        json={"cn", uid},
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

        client.connect("18.188.110.246", 8883)
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

