from config import supabase 
from utils import error_response, success_response

def get_devices(uid):
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