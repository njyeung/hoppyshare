from mosquitto_api import pub_settings
from config import supabase 

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

    res = pub_settings(new_settings, uid)

    return res