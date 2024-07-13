from flask import Flask, send_from_directory, request, abort, jsonify
import os
import base64
import bcrypt

app = Flask(__name__)

# Directory paths
FIRMWARE_DIR = '/data/firmware'
CONFIGURATIONS_DIR = '/data/configurations'
AUTH_DIR = '/data/auth'

def check_auth(device, password):
    """Check if the password matches the hash stored for the device."""
    hash_file = os.path.join(AUTH_DIR, f"{device}.hash")
    if os.path.exists(hash_file):
        with open(hash_file, 'rb') as file:
            stored_hash = file.read().strip()
        return bcrypt.checkpw(password.encode('utf-8'), stored_hash)
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

@app.route('/firmware/<path:filename>')
def serve_firmware(filename):
    """Serve firmware files."""
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

