import os
import json

def build_binary(platform, device_id, cert, key, group_key):
    output_dir = "./tmp/output"
    os.makedirs(output_dir, exist_ok=True)

    # Store platform targets
    targets = {
        "LINUX":     "device_linux",
        "MACOS":     "device_darwin",
        "WINDOWS":   "device_windows.exe"
    }

    if platform not in targets:
        print(f"[build_binary] error: unsupported platform")
        return None

    binary_path = os.path.abspath(f"./tmp/{targets[platform]}")
    if not os.path.exists(binary_path):
        print(f"[build_binary] error: prebuilt binary not found: {binary_path}")
        return None

    # Load CA cert
    ca_cert = open("./tmp/ca.crt").read()

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

    # Write files out (dev)
    output_binary_path = os.path.join(output_dir, "out")
    with open(output_binary_path, "wb") as f:
        f.write(final_binary)
    os.chmod(output_binary_path, 0o755)
    
    metadata_json_path = os.path.join(output_dir, f"out.metadata.json")
    with open(metadata_json_path, "w") as f:
        json.dump(metadata, f, indent=2)

    return final_binary