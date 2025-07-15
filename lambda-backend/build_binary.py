import os
import subprocess

def build_binary(platform, device_id, cert, key, group_key):
    output_dir = "./tmp/desktop_client/mqttclient/lambda_output"

    os.makedirs(output_dir, exist_ok=True)

    with open(f"{output_dir}/cert.pem", "w") as f:
        f.write(cert)
    with open(f"{output_dir}/key.pem", "w") as f:
        f.write(key)
    with open(f"{output_dir}/ca.crt", "w") as f:
        f.write(open("./certs/ca.crt").read())
    with open(f"{output_dir}/group_key.enc", "wb") as f:
        f.write(group_key)
    with open(f"{output_dir}/device_id.txt", "w") as f:
        f.write(device_id)
    
    targets = {
        "linux": ("linux", "amd64", "snap_notes_device"), 
        "macos": ("darwin", "amd64", "snap_notes_device"), 
        "windows": ("windows", "amd64", "snap_notes_device.exe")
    }

    if platform not in targets:
        print(f"[build_binary] error: platform not in targets")
        return None

    goos, goarch, output_name = targets[platform]

    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch

    build_path = os.path.abspath(f"./tmp/snap_notes_device")
    
    result = subprocess.run(
        ["go", "build", "-o", build_path, "."],
        cwd="./tmp/desktop_client",
        capture_output=True,
        text=True,
        env=env
    )

    if result.returncode != 0:
        print(f"[build_binary] error: {result}")
        return None

    # Send binary as file or base64
    with open("./tmp/device_client", "rb") as f:
        binary_data = f.read()

    return binary_data