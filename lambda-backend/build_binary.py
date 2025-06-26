import os
import subprocess

def build_binary(cert, key):

        os.makedirs("./tmp/lambda_output", exist_ok=True)

        with open("./tmp/lambda_output/cert.pem", "w") as f:
            f.write(cert)
        with open("./tmp/lambda_output/key.pem", "w") as f:
            f.write(key)
        with open("./tmp/lambda_output/ca.crt", "w") as f:
            f.write(open("./certs/ca.crt").read())

        result = subprocess.run(
            ["go", "build", "-o", "./device_client", "./client.go"],
            cwd="./tmp",
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            return {"error": result.stderr}

        # Send binary as file or base64
        with open("./tmp/device_client", "rb") as f:
            binary_data = f.read()

        return binary_data