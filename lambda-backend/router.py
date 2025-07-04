from utils import success_response, error_response, get_uid_from_auth_header, forbidden_response
from config import SUPABASE_SERVICE_SECRET
from actions.onboard_user import onboard_user
from actions.add_device import add_device
from actions.get_devices import get_devices
from actions.revoke_device import revoke_device
from actions.change_settings import change_settings

def route_action(headers, body):
    action = body.get("action")
    
    # Route for supabase webhook
    match action:
        case "onboard_user":
            auth_header = headers.get("Authorization")
            if auth_header != f"Bearer {['SUPABASE_SERVICE_SECRET']}":
                return forbidden_response("Invalid service token")
            
            uid = body.get("uid")
            if not uid:
                return error_response("Missing uid")

            return onboard_user(uid)


    # Protected routes (using supabase jwt)
    try:
        uid = get_uid_from_auth_header(headers)
    except Exception as e:
        return forbidden_response(str(e))

    match action:
        case "add_device":
            return add_device(uid)
        case "get_devices":
            return get_devices(uid)
        case "revoke_device":
            cert = body.get("cert")
            
            if not cert:
                return error_response("Cert field required")

            return revoke_device(uid, cert)
        case "change_settings":
            device_id = body.get("device_id")
            new_settings = body.get("new_settings")
             
            if not device_id:
                return error_response("device_id field requried")
            if not new_settings:
                return error_response("new_settings field required")

            return change_settings(uid, device_id, new_settings)
    
    return error_response("Unknown action")