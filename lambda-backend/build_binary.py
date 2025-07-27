import os
import json
import boto3
import base64
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

def generate_key():
    return AESGCM.generate_key(bit_length=256)  # 32 bytes

def encrypt_blob(key: bytes, plaintext_dict: dict, aad: bytes) -> str:
    aesgcm = AESGCM(key)
    nonce = os.urandom(12)
    plaintext = json.dumps(plaintext_dict).encode("utf-8")
    ciphertext = aesgcm.encrypt(nonce, plaintext, aad)
    return base64.b64encode(nonce + ciphertext).decode("utf-8")

def build_binary(platform, device_id, cert, key, group_key):
    s3 = boto3.client("s3")
    output_dir = "/tmp/output"
    os.makedirs(output_dir, exist_ok=True)

    # Define S3 locations for precompiled base binaries
    targets = {
        "LINUX":     "device_linux",
        "MACOS":     "device_darwin",
        "WINDOWS":   "device_windows.exe"
    }

    if platform not in targets:
        print(f"[build_binary] error: unsupported platform")
        return None

    bucket_name = "hoppyshare-binaries"
    object_key = targets[platform]

    try:
        print(f"[build_binary] downloading {object_key} from s3://{bucket_name}")
        response = s3.get_object(Bucket=bucket_name, Key=object_key)
        base_binary = response["Body"].read()
    except Exception as e:
        print(f"[build_binary] error downloading from S3: {e}")
        return None

    # Load CA cert
    try:
        ca_cert = open("./certs/ca.crt").read()
    except Exception as e:
        print(f"[build_binary] error loading CA cert: {e}")
        return None

    # Encrypt certs and keys
    encryption_key = generate_key()
    encryption_key_b64 = base64.b64encode(encryption_key).decode("utf-8")

    secret_payload = {
        "cert": cert,
        "key": key,
        "ca_cert": ca_cert,
        "group_key": group_key.hex()
    }

    aad = device_id.encode("utf-8")
    encrypted_blob_b64 = encrypt_blob(encryption_key, secret_payload, aad)

    # Create the embedded metadata
    metadata = {
        "device_id": device_id,
        "encrypted_blob": encrypted_blob_b64
    }

    marker = b"\n--APPEND_MARKER--\n"
    json_blob = json.dumps(metadata).encode("utf-8")
    final_binary = base_binary + marker + json_blob

    # Debug output to /tmp
    try:
        output_binary_path = os.path.join(output_dir, "out")
        with open(output_binary_path, "wb") as f:
            f.write(final_binary)
        os.chmod(output_binary_path, 0o755)

        metadata_json_path = os.path.join(output_dir, "out.metadata.json")
        with open(metadata_json_path, "w") as f:
            json.dump(metadata, f, indent=2)

    except Exception as e:
        print(f"[build_binary] warning: could not write debug files: {e}")

    return final_binary, encryption_key_b64
