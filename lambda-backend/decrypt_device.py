import base64
import json
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from utils import error_response, success_response
from config import supabase

def decrypt_device(device_id, body):
    encrypted_blob_b64 = body.get("encrypted_blob")
    if not encrypted_blob_b64:
        return error_response("Missing encrypted_blob in request body")

    # Fetch encryption key from Supabase
    try:
        res = supabase.table("encryption_keys").select("encryption_key, used").eq("deviceid", device_id).single().execute()
        data = res.data
        if not data:
            return error_response("Encryption key not found for device")

        # Allow for now
        # if data.get("used"):
        #     return error_response("Encryption key has already been used")
        
        key = base64.b64decode(data["encryption_key"])
    except Exception as e:
        return error_response("Failed to fetch encryption key")

    # Attempt to decrypt the blob
    try:
        blob_bytes = base64.b64decode(encrypted_blob_b64)
        nonce = blob_bytes[:12]
        ciphertext = blob_bytes[12:]

        aesgcm = AESGCM(key)
        decrypted = aesgcm.decrypt(nonce, ciphertext, device_id.encode())
    except Exception as e:
        return error_response("Failed to decrypt blob. Invalid encrypted_blob")

    # Mark the encryption key as used
    try:
        supabase.table("encryption_keys").update({"used": True}).eq("deviceid", device_id).execute()
    except Exception as e:
        return error_response("Failed to mark encryption key as used", str(e))

    # Everything checks out, return the encryption key
    return success_response({
        "encryption_key": data["encryption_key"]
    })
