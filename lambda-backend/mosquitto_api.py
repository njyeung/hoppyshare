import requests
import os
import subprocess
import json
from build_binary import build_binary
import base64

MOSQUITTO_API = "https://localhost"
CERT = "./certs/lambda.crt"
KEY = "./certs/lambda.key"
CA = "./certs/ca.crt"

def onboard_user(uid):
    return requests.post(
        f"{MOSQUITTO_API}/onboard_user",
        json={"cn": uid},
        cert=(CERT, KEY),
        verify=CA
    )

def add_device(uid):
    response = requests.post(
        f"{MOSQUITTO_API}/add_device",
        json={"cn": uid},
        cert=(CERT, KEY),
        verify=CA
    )

    if response.status_code != 200:
        return {"error": f"Failed to add device: {response.text}"}

    cert = response.json().get("cert")
    key = response.json().get("key")

    binary = build_binary(cert, key)

    if binary is None:
        return {"error": "Failed to build device binary"}

    encoded = base64.b64encode(binary).decode("utf-8")

    return {
        "binary": encoded
    }

def revoke_device(cert_pem):
    return requests.post(
        f"{MOSQUITTO_API}/revoke_device",
        json={"cert": cert_pem},
        cert=(CERT, KEY),
        verify=CA
    )
