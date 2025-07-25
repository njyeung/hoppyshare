from mosquitto_api import pub_settings
from config import supabase 
from get_devices import get_devices
from utils import error_response, forbidden_response, success_response
def change_settings(uid, device_id, new_settings):

    # Make sure user owns this device
    try:
        query = (
            supabase.table("device")
            .select("deviceid")
            .eq("uid", uid)
            .execute()
        )
        devices = query.data

        if not any(d["deviceid"] == device_id for d in devices):
            return forbidden_response("Device does not belong to user")
    except Exception as e:
        return error_response("Failed to query devices", str(e))

    # Change settings in database
    update_res = ( supabase.table("device")
        .update({"settings": new_settings})
        .eq("deviceid", device_id)
        .execute()
    )

    # Get new settings list
    query = get_devices(uid)
    
    if query.get("status_code") != 200:
        return error_response("Failed to change settings")

    settings = query.get("json", {}).get("devices", [])

    res = pub_settings(settings, uid)

    return res