from mosquitto_api import onboard_user, add_device, revoke_device
import json

def route_action(body):
    action = body.get("action")

    if action == "onboard_user":
        uid = body.get("uid")
        if not uid:
            return {"statusCode": 400, "body": "Missing uid"}
        res = onboard_user(uid)

    elif action == "add_device":
        uid = body.get("uid")
        if not uid:
            return {"statusCode": 400, "body": "Missing uid"}
        res = add_device(uid)

    elif action == "revoke_device":
        cert = body.get("cert")
        if not cert:
            return {"statusCode": 400, "body": "Missing cert"}
        res = revoke_device(cert)

    else:
        return {"statusCode": 400, "body": f"Unknown action '{action}'"}

    if isinstance(res, dict):
        # If it's an error dict
        if "error" in res:
            return {"statusCode": 500, "body": res["error"]}
        else:
            return {"statusCode": 200, "body": json.dumps(res)}
            
    return {
        "statusCode": res.status_code,
        "body": json.dumps({"response": res.json()})
    }
