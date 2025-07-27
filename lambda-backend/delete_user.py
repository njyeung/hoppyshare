from mosquitto_api import delete_user
from config import SUPABASE_URL, SUPABASE_SERVICE_SECRET
from utils import success_response, error_response

def delete_user(uid):
    # Delete user from supabase
    url = f"{SUPABASE_URL}/auth/v1/admin/users/{uid}"
    headers = {
        "apikey": SUPABASE_SERVICE_SECRET,
        "Authorization": f"Bearer {SUPABASE_SERVICE_SECRET}"
    }

    res = requests.delete(url, headers=headers)
    
    if res.status_code == 204:
        return {"status": "User deleted from Supabase auth"}
    else:
        return {
            "error": "Failed to delete user from Supabase auth",
            "status_code": res.status_code,
            "details": res.text
        }
    
    # Remove their ACL from mosquitto
    res = delete_user(uid)

    if res.get("status_code") != 200:
        return error_response("Could not delete user from mosquitto, likely because they were already deleted")
    
    return success_response(f"Deleted user {uid}")


    