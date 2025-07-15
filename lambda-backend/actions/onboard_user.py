import os
from config import supabase
import mosquitto_api
from utils import success_response, error_response

def generate_group_key():
    return os.urandom(32)  # AES-256

def onboard_user(uid: str):
    # Call Mosquitto to set up topics and ACLs
    res = mosquitto_api.onboard_user(uid)

    if res["status_code"] != 200:
        return error_response("Failed to onboard user", res)

    group_key = generate_group_key()
    hex_key = "\\x" + group_key.hex()

    # Store the key in Supabase for E2EE
    res = supabase.table("user_keys").insert({
        "user_id": uid,
        "group_key": hex_key
    }).execute()

    return success_response("ok")