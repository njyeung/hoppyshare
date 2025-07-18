import os
import json

def build_binary(platform: PLATFORM, device_id, cert, key, group_key):
    """
    Appends device-specific credentials to a prebuilt binary for the given platform.
    Returns the final binary as bytes.
    """
    # Paths
    output_dir = "./tmp/output"
    os.makedirs(output_dir, exist_ok=True)

    # Store platform targets
    targets = {
        PLATFORM.LINUX:     "device_linux",
        PLATFORM.MACOS:     "device_darwin",
        PLATFORM.WINDOWS:   "device_windows.exe"
    }

    if platform not in targets:
        print(f"[build_binary] error: unsupported platform")
        return None

    binary_path = os.path.abspath(f"./tmp/{targets[platform]}")
    if not os.path.exists(binary_path):
        print(f"[build_binary] error: prebuilt binary not found: {binary_path}")
        return None

    # Load CA cert
    ca_cert = open("./certs/ca.crt").read()

    # Bundle credentials + metadata
    metadata = {
        "device_id": device_id,
        "cert": cert,
        "key": key,
        "ca_cert": ca_cert,
        "group_key": group_key.hex()
    }

    # Append to binary
    marker = b"\n--APPEND_MARKER--\n"
    json_blob = json.dumps(metadata).encode("utf-8")
    final_binary = open(binary_path, "rb").read() + marker + json_blob

    return final_binary