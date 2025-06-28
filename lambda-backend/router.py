from mosquitto_api import onboard_user, add_device, revoke_device, pub_settings
from utils import success_response, error_response, get_uid_from_auth_header, forbidden_response
from build_binary import build_binary
import base64
from supabase import create_client, Client
import uuid
import json
import os
from dotenv import load_dotenv

load_dotenv()

SUPABASE_URL = os.environ.get("SUPABASE_URL")
SUPABASE_KEY = os.environ.get("SUPABASE_KEY")

supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)

def route_action(headers, body):
    uid = get_uid_from_auth_header(headers)
    action = body.get("action")

    if action == "onboard_user":
        return onboard_user(uid)

    elif action == "add_device":
        if not uid:
            return forbidden_response("Not authorized")

        # Add device to mosquitto
        res = add_device(uid)
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
            return error_response("Failed to insert device into database", str(e))

        # Build binary and return it
        binary = build_binary(device_id, cert, key)
        if binary is None:
            return error_response("Failed to build device binary")

        encoded = base64.b64encode(binary).decode("utf-8")
        return success_response({"binary": encoded})

    elif action == "get_devices":
        if not uid:
            return forbidden_response("Not authorized")

        try:
            query = (
                supabase.table("device")
                .select("deviceid, settings")
                .eq("uid", uid)
                .execute()
            )
            devices = query.data
            return success_response({"devices": devices})

        except Exception as e:
            return error_response("Failed to query devices", str(e))

    elif action == "revoke_device":
        if not uid:
            return forbidden_response("Not authorized")
        cert = body.get("cert")
        if not cert:
            return error_response("Missing cert")
        return revoke_device(cert)

    elif action == "change_settings":
        if not uid:
            return forbidden_response("Not authorized")

        device_id = body.get("device_id")
        new_settings = body.get("new_settings")

        res = pub_settings({"copy": True}, "user1")
        return res

    else:
        return error_response(f"Unknown action '{action}'")
