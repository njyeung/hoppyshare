import os
import subprocess

def build_binary(platform, device_id, cert, key):
    os.makedirs("./tmp/lambda_output", exist_ok=True)

    with open("./tmp/lambda_output/cert.pem", "w") as f:
        f.write(cert)
    with open("./tmp/lambda_output/key.pem", "w") as f:
        f.write(key)
    with open("./tmp/lambda_output/ca.crt", "w") as f:
        f.write(open("./certs/ca.crt").read())
    with open("./tmp/lambda_output/device_id.txt", "w") as f:
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

    result = subprocess.run(
        ["go", "build", "-o", f"./{output_name}", "./client.go"],
        cwd="./tmp",
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