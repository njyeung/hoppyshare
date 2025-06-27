# router.py

from mosquitto_api import onboard_user, add_device, revoke_device, pub_settings
from build_binary import build_binary
import base64
from utils import success_response, error_response

def route_action(body):
    action = body.get("action")
    uid = body.get("uid")

    if action == "onboard_user":
        if not uid:
            return error_response("Missing uid")
        return onboard_user(uid)

    elif action == "add_device":
        if not uid:
            return error_response("Missing uid")

        res = add_device(uid)
        if res["status_code"] != 200:
            return error_response("Failed to add device", res)

        cert = res["json"].get("cert")
        key = res["json"].get("key")
        if not cert or not key:
            return error_response("Cert or key missing in response")

        binary = build_binary(cert, key)
        if binary is None:
            return error_response("Failed to build device binary")

        encoded = base64.b64encode(binary).decode("utf-8")
        return success_response({"binary": encoded})

    elif action == "revoke_device":
        cert = body.get("cert")
        if not cert:
            return error_response("Missing cert")
        return revoke_device(cert)

    elif action == "change_settings":
        res = pub_settings({"copy": True}, "user1")
        return res

    else:
        return error_response(f"Unknown action '{action}'")
