import requests
from functools import wraps
from jose import jwt, JWTError
import os

SUPABASE_JWT_SECRET = os.environ.get("SUPABASE_JWT_SECRET")

if not SUPABASE_JWT_SECRET:
    raise ValueError("Missing required environment variable: SUPABASE_JWT_SECRET")

def get_uid_from_auth_header(headers):
    auth_header = headers.get("Authorization", "")
    if not auth_header.startswith("Bearer "):
        raise Exception("Missing or malformed Authorization header")

    token = auth_header[len("Bearer "):]
    try:
        payload = jwt.decode(
            token, 
            SUPABASE_JWT_SECRET, 
            algorithms=["HS256"], 
            audience="authenticated")
        
        return payload["sub"]

    except JWTError as e:
        raise Exception(f"JWT verification failed: {e}")

def wrap_response(response: requests.Response):
    try:
        body = response.json()
    except Exception:
        body = {"raw": response.text}
    return {
        "status_code": response.status_code,
        "json": body
    }

def api_response(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        res = fn(*args, **kwargs)
        if isinstance(res, requests.Response):
            return wrap_response(res)
        return res
    return wrapper


def error_response(message, original=None):
    return {
        "status_code": original["status_code"] if original else 400,
        "json": {
            "error": message,
            "details": original["json"] if original else None
        }
    }

def success_response(data):
    return {
        "status_code": 200,
        "json": data
    }

def forbidden_response(message):
    return {
        "status_code": 403,
        "json": {
            "error": message,
        }
    }