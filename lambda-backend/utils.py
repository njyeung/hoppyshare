# utils.py

import requests
from functools import wraps

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
