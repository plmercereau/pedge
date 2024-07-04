import os
import csv

csv_file_path = os.getenv('CSV_FILE')

with open(csv_file_path, 'w', newline='') as csvfile:
    writer = csv.writer(csvfile, delimiter=',', quotechar='"', quoting=csv.QUOTE_MINIMAL)
    writer.writerow(['key','type','encoding','value'])

    writer.writerow(['mqtt','namespace','',''])
    # TODO MQTT settings from environment variables

    # TODO custom settings from /secrets

        

