from utils import success_response, error_response, get_uid_from_auth_header, forbidden_response
from config import SUPABASE_SERVICE_SECRET
from onboard_user import onboard_user
from add_device import add_device
from get_devices import get_devices
from revoke_device import revoke_device
from change_settings import change_settings
from delete_user import delete_user
from decrypt_device import decrypt_device
import json

def route_action(event):
    method = event["httpMethod"]
    path = event["resource"]
    headers = event.get("headers", {})
    body = json.loads(event["body"]) if event.get("body") else {}
    pathParameters = event.get("pathParameters", {})

    # Debug: Log all incoming requests
    print(f"DEBUG - Method: {method}, Path: {path}")
    print(f"DEBUG - Full event: {json.dumps(event, indent=2)}")

    # Route for supabase webhook
    if method == "POST" and path == "/api/onboard":
        print(f"Webhook received - Method: {method}, Path: {path}")
        print(f"All headers: {headers}")
        print(f"Body: {body}")
        
        auth_header = headers.get("Auth") or headers.get("auth")
        print(f"Auth header: {auth_header}")
        print(f"Expected: {SUPABASE_SERVICE_SECRET}")
        
        if auth_header != SUPABASE_SERVICE_SECRET:
            return forbidden_response("Invalid service token")
        
        # Supabase sends user data in webhook payload
        user_data = body.get("record", {})
        uid = user_data.get("id")
        
        if not uid:
            return error_response("Missing user ID in webhook payload")
        
        return onboard_user(uid)

    # Public routes

    match(method, path):
        case ("POST", "/api/decrypt/{device_id}"):
            device_id = pathParameters.get("device_id", None)
            if not device_id:
                return error_response("Missing device_id in path")
            
            return decrypt_device(device_id, body)

    # Protected routes (using supabase jwt)

    try:
        uid = get_uid_from_auth_header(headers)
    except Exception as e:
        return forbidden_response(str(e))

    match (method, path):
        case ("POST", "/api/onboard"):
            return onboard_user(uid)
        case ("POST" ,"/api/devices"):
            platform = body.get("platform")

            if not platform:
                return error_response("platform field required")

            if platform not in { "WINDOWS", "MACOS", "LINUX" }:
                return error_response("platform must be LINUX, MACOS, or WINDOWS") 

            return add_device(uid, platform)
        case ("GET", "/api/devices"):
            return get_devices(uid)
        case ("DELETE", "/api/devices/{device_id}"):
            device_id = pathParameters.get("device_id", None)
            
            if not device_id:
                return error_response("device_id field required")

            return revoke_device(uid, device_id)
        case ("PUT", "/api/settings/{device_id}"):
            device_id = pathParameters.get("device_id", None)
            new_settings = body.get("new_settings", None)
             
            if not device_id:
                return error_response("device_id field requried")
            if not new_settings:
                return error_response("new_settings field required")

            return change_settings(uid, device_id, new_settings)
        case ("DELETE", "/api/user"):
            return delete_user(uid)

    return error_response("Unknown endpoint")