import subprocess
from flask import Flask, jsonify, request
import tempfile
import os
import datetime

CA_KEY = "/mosquitto/certs/ca.key"
CA_CERT = "/mosquitto/certs/ca.crt"
OPENSSL_CONF = "/mosquitto/certs/openssl.cnf"
CERT_DIR = "/mosquitto/certs"

app = Flask(__name__)

@app.route('/stop', methods=['POST'])
def stop_mosquitto():
    try:
        subprocess.run(["pkill", "mosquitto"], check=True)
        return jsonify({"status": "stopped"}), 200
    except subprocess.CalledProcessError:
        return jsonify({"error": "Failed to stop mosquitto"}), 500

@app.route('/start', methods=['POST'])
def start_mosquitto():
    try:
        subprocess.Popen(["/usr/sbin/mosquitto", "-c", "/mosquitto/config/mosquitto.conf"])
        return jsonify({"status": "started"}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/onboard_user', methods=['POST'])
def onboard_user():
    data = request.get_json()
    cn = data.get("cn")
    if not cn:
        return jsonify({"error": "Missing CN"}), 400

    acl_file_path = os.path.join(DYNAMIC_DIR, f"user_{cn}.acl")

    if os.path.exists(acl_file_path):
        return jsonify({"error": "User already onboarded"}), 400

    try:
        with open(acl_file_path, "w") as f:
            f.write(f"""user {cn}
            topic write users/{cn}/notes
            topic read users/{cn}/notes
            topic read users/{cn}/settings
            """)

        # Secure the file
        os.chmod(acl_file_path, 0o600)

        # Call /reload internally
        reload_response = reload_mosquitto()
        if reload_response[1] != 200:
            return jsonify({"error": "User ACL created, but reload failed"}), 500

        return jsonify({"status": f"user {cn} onboarded"}), 200

    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/add_device', methods=['POST'])
def add_device():
    data = request.get_json()
    cn = data.get("cn")
    if not cn:
        return jsonify({"error": "Missing CN"}), 400
    
    acl_file_path = os.path.join(DYNAMIC_DIR, f"user_{cn}.acl")
    if not os.path.exists(acl_file_path):
        return jsonify({"error": "User CN does not exist"}), 400

    timestamp = datetime.datetime.utcnow().strftime("%Y%m%d%H%M%S")
    base = os.path.join(CERT_DIR, f"{cn}_{timestamp}")
    key_path = f"{base}.key"
    csr_path = f"{base}.csr"
    crt_path = f"{base}.crt"

    try:
        # Generate private key
        subprocess.run(["openssl", "genrsa", "-out", key_path, "2048"], check=True)

        # Generate CSR
        subprocess.run([
            "openssl", "req", "-new", "-key", key_path,
            "-out", csr_path, "-subj", f"/CN={cn}"
        ], check=True)

        # Sign cert
        subprocess.run([
            "openssl", "x509", "-req", "-in", csr_path,
            "-CA", CA_CERT, "-CAkey", CA_KEY, "-CAcreateserial",
            "-out", crt_path, "-days", "365", "-sha256",
            "-extfile", OPENSSL_CONF, "-extensions", "v3_req"
        ], check=True)

        # Read contents
        with open(crt_path) as f:
            cert_pem = f.read()
        with open(key_path) as f:
            key_pem = f.read()
        
        # Clean up files
        for path in [key_path, csr_path, crt_path]:
            if os.path.exists(path):
                os.remove(path)
        
        return jsonify({
            "cert": cert_pem,
            "key": key_pem
        }), 200

    except subprocess.CalledProcessError as e:
        return jsonify({"error": f"OpenSSL error: {e}"}), 500
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/delete_user', methods=['POST'])
def delete_user():
    data = request.get_json()
    cn = data.get("cn")
    if not cn:
        return jsonify({"error": "Missing CN"}), 400

    try:
        # Delete all matching .acl files for this CN
        deleted = []
        for fname in os.listdir(DYNAMIC_DIR):
            if fname.startswith(f"user_{cn}") and fname.endswith(".acl"):
                full_path = os.path.join(DYNAMIC_DIR, fname)
                os.remove(full_path)
                deleted.append(fname)

        if not deleted:
            return jsonify({"error": f"No ACL file found for CN={cn}"}), 404

        # Trigger reload
        reload_response = reload_mosquitto()
        if reload_response[1] != 200:
            return jsonify({"error": "ACL file removed, but reload failed"}), 500

        return jsonify({"status": f"Revoked device(s) for CN={cn}", "files_removed": deleted}), 200

    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/revoke_device', methods=['POST'])
def revoke_device():
    data = request.get_json()
    cert_pem = data.get("cert")
    if not cert_pem:
        return jsonify({"error": "Missing device certificate"}), 400

    try:
        # Save incoming cert to a temporary file
        with tempfile.NamedTemporaryFile(delete=False, suffix=".crt") as temp_cert:
            temp_cert.write(cert_pem.encode("utf-8"))
            cert_path = temp_cert.name

        # Revoke the cert
        cmd_revoke = ["openssl", "ca", "-config", "/mosquitto/certs/openssl.cnf", "-revoke", cert_path]
        result = subprocess.run(cmd_revoke, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if result.returncode != 0:
            return jsonify({"error": "Failed to revoke cert", "details": result.stderr.decode()}), 500

        # Regenerate CRL
        cmd_crl = ["openssl", "ca", "-gencrl", "-out", "/mosquitto/certs/ca-crl.pem", "-config", "/mosquitto/certs/openssl.cnf"]
        result = subprocess.run(cmd_crl, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if result.returncode != 0:
            return jsonify({"error": "Failed to generate CRL", "details": result.stderr.decode()}), 500

        # Reload Mosquitto
        reload_response = reload_mosquitto()
        if reload_response[1] != 200:
            return jsonify({"error": "Cert revoked but Mosquitto reload failed"}), 500

        return jsonify({"status": "Device cert revoked"}), 200

    except Exception as e:
        return jsonify({"error": str(e)}), 500

BASE_ACL = "/mosquitto/baseACL"
DYNAMIC_DIR = "/mosquitto/dynamic_acl"
MERGED_ACL = "/mosquitto/config/merged_acl.txt"

@app.route('/reload', methods=['POST'])
def reload_mosquitto():
    try:
        dynamic_sections = []
        pattern_sections = []

        # Separate baseACL into user and pattern sections
        with open(BASE_ACL, "r") as f:
            current_block = []
            in_pattern_block = False
            for line in f:
                stripped = line.strip()
                if stripped.startswith("pattern"):
                    in_pattern_block = True
                elif stripped.startswith("user"):
                    in_pattern_block = False

                if in_pattern_block:
                    pattern_sections.append(line)
                else:
                    current_block.append(line)

            user_sections = "\n".join(current_block)

        # Read all dynamic ACL blocks first (deny rules etc.)
        for fname in sorted(os.listdir(DYNAMIC_DIR)):
            if fname.endswith(".acl"):
                with open(os.path.join(DYNAMIC_DIR, fname), "r") as f:
                    dynamic_sections.append(f.read())

        # Compose merged ACL
        with open(MERGED_ACL, "w") as f:
            f.write("# --- User-specific dynamic blocks (deny rules) ---\n")
            f.write("\n\n".join(dynamic_sections))
            f.write("\n\n# --- User rules from baseACL ---\n")
            f.write(user_sections)
            f.write("\n\n# --- Pattern rules from baseACL ---\n")
            f.write("".join(pattern_sections))

        # Fix file permissions if needed
        os.chmod(MERGED_ACL, 0o600)

        # Reload Mosquitto
        subprocess.run(["pkill", "-HUP", "mosquitto"], check=True)

        return jsonify({"status": "merged and reloaded"}), 200

    except Exception as e:
        return jsonify({"error": str(e)}), 500

