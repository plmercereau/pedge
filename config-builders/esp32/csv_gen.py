import os
import csv

with open(os.getenv('TMP_CSV_FILE'), 'w', newline='') as csvfile:
    writer = csv.writer(csvfile, delimiter=',')
    writer.writerow(['key','type','encoding','value'])

    def start_namespace(namespace):
        print("Namespace:", namespace)
        writer.writerow([namespace,'namespace','',''])
    
    def add_string_data(key, value):
        print(key, "(empty value)" if not value else "")
        writer.writerow([key,'data','string',value])
    
    start_namespace('mqtt')
    add_string_data('BROKER', os.getenv('MQTT_BROKER'))
    add_string_data('PORT', os.getenv('MQTT_PORT'))
    add_string_data('USERNAME', os.getenv('MQTT_USERNAME'))
    add_string_data('PASSWORD', os.getenv('MQTT_PASSWORD'))
    add_string_data('SENSORS_TOPIC', os.getenv('MQTT_SENSORS_TOPIC'))
    
    secretsPath = os.getenv('SECRETS_PATH', '/secrets')
    settings = {}
    reserved_keys = ['config.bin', 'username', 'password']
    if os.path.exists(secretsPath):
        for file in os.listdir(secretsPath):
            # add to the settings if file is a file and its name is not in reserved_keys
            file_path = secretsPath + '/' + file
            if os.path.isfile(file_path) and file not in reserved_keys:
                with open(file_path, 'r') as f:
                    settings[file] = f.read()

    if settings:
        start_namespace('settings')
        for key, value in settings.items():
            add_string_data(key, value)

    
        

