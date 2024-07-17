from flask import Flask, send_from_directory, request, abort, jsonify
import os
import base64
from argon2 import PasswordHasher

app = Flask(__name__)

# Directory paths
FIRMWARE_DIR = '/data/firmwares'
CONFIGURATIONS_DIR = '/data/configurations'
AUTH_DIR = '/data/auth'

ph = PasswordHasher()

def check_auth(device, password):
    """Check if the password matches the hash stored for the device."""
    hash_file = os.path.join(AUTH_DIR, f"{device}.hash")
    if os.path.exists(hash_file):
        with open(hash_file, 'rb') as file:
            stored_hash = file.read().strip()
            try:
                ph.verify(stored_hash, password)
                return True
            except:
                return False
    return False

def authenticate():
    """Sends a 401 response that enables basic auth"""
    return abort(401, 'Unauthorized access. Please provide valid credentials.')

def requires_auth(f):
    """Decorator to prompt for basic auth and check credentials."""
    def decorated(*args, **kwargs):
        auth = request.authorization
        if not auth or not check_auth(kwargs['device'], auth.password):
            return authenticate()
        return f(*args, **kwargs)
    return decorated

@app.route('/firmwares/<path:filename>')
def serve_firmware(filename):
    """Serve firmware files."""
    app.logger.debug(f"Request for firmware: {filename}")
    return send_from_directory(FIRMWARE_DIR, filename)

@app.route('/configurations/<device>/<path:filename>')
@requires_auth
def serve_configuration(device, filename):
    """Serve configuration files with basic auth."""
    return send_from_directory(os.path.join(CONFIGURATIONS_DIR, device), filename)

@app.route('/health', methods=['GET'])
def health_check():
    response = {
        "status": "healthy",
        "message": "The application is running smoothly."
    }
    return jsonify(response), 200

