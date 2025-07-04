import mosquitto_api

def revoke_device(uid, cert):
    return mosquitto_api.revoke_device(cert)