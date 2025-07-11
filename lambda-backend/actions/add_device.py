from mosquitto_api import pub_settings, add_device
import mosquitto_api
from config import supabase 
from actions.get_devices import get_devices
from build_binary import build_binary
import uuid
from utils import error_response, success_response
import base64
import json

def add_device(uid):
    # Add device to mosquitto
    res = mosquitto_api.add_device(uid)
    
    if res["status_code"] != 200:
        return error_response("Failed to add device", res)
    cert = res["json"].get("cert")
    key = res["json"].get("key")

    if not cert or not key:
        return error_response("Cert or key missing in response")

    # Add device to DB
    device_id = str(uuid.uuid4())
    data = {
        "deviceid": device_id,
        "uid": uid,
        "settings": {"copy": True}, 
        "cert": cert
    }

    try:
        insert_res = supabase.table("device").insert(data).execute()
    except Exception as e:
        return error_response("Failed to insert device into database")
    
    # Fetch existing settings
    query = get_devices(uid)
    
    if query.get("status_code") != 200:
        return error_response("Failed to set up device settings")
    
    settings = query.get("json", {}).get("devices", [])

    # Set up settings
    res = pub_settings(settings, uid)

    print(res)

    if res.get("status_code")!= 200:
        return error_response("Failed to set up device settings")

    # Build binary and return it
    binary = build_binary("linux", device_id, cert, key)
    if binary is None:
        return error_response("Failed to build device binary")

    encoded = base64.b64encode(binary).decode("utf-8")
    return success_response({"binary": encoded})
