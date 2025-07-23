import mosquitto_api
from config import supabase
from utils import error_response, forbidden_response
from get_devices import get_devices

def revoke_device(uid, device_id):
    try:
        # Check if device belongs to user and get its cert
        query = (
            supabase.table("device")
            .select("cert")
            .eq("deviceid", device_id)
            .eq("uid", uid)
            .execute()
        )

        data = query.data
        
        if not data:
            return forbidden_response("Device id does not exist or does not belong to user")
        
        cert = data[0]["cert"]

        # Revoke the cert in mosquitto
        res = mosquitto_api.revoke_device(cert)

        if res.get("status_code") != 200:
            return error_response("Could not revoke cert, likely because it was already revoked")
        
        # Remove device from database
        delete_res = (
            supabase.table("device")
            .delete()
            .eq("deviceid", device_id)
            .eq("uid", uid)
            .execute()
        )

        # Update settings topic
        query = get_devices(uid)

        if query.get("status_code") != 200:
            return error_response("Failed to update settings")

        settings = query.get("json", {}).get("devices", [])

        pub_res = mosquitto_api.pub_settings(settings, uid)

        if pub_res.get("status_code") != 200:
            return error_response("Could not publish new settings to topic")

        return res
        
    except Exception as e:
        return error_response("Failed to revoke device", str(e))
    
    