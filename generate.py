import csv
import re
import json

user_prompts_fields = ['request_id', 'user_id', 'prompt', 'timestamp']
openai_response_fields = ['request_id', 'user_id', 'response', 'timestamp']

user_prompts_data = []
openai_response_data = []

with open('/Users/newsales/dev-projects/diagramastext/logs/lambda/logs') as log_file:
    for line in log_file:
        request_id_match = re.search(r'START RequestId: ([\w-]+)', line)
        if request_id_match:
            request_id = request_id_match.group(1)
        prompt_match = re.search(r'{\s*"prompt":\s*"([^"]+)"\s*}', line)
        if prompt_match:
            timestamp = line.split()[0]
            prompt = prompt_match.group(1)
            user_prompts_data.append([request_id, '00000000-0000-0000-0000-000000000000', prompt, timestamp])
        
        if 'cleaned' in line:
            try:
                timestamp = line.split()[0]
                cleaned_value_start = line.find('{"cleaned":')  # find the start of the cleaned value
                cleaned_value_end = line.rfind('}')  # find the end of the cleaned value
                response = line[cleaned_value_start+12:cleaned_value_end+1]
                openai_response_data.append([request_id, '00000000-0000-0000-0000-000000000000', response, timestamp])
            except:
                pass
            
with open('user_prompts.csv', 'w', newline='') as user_prompts_file:
    csv_writer = csv.writer(user_prompts_file)
    csv_writer.writerow(user_prompts_fields)
    csv_writer.writerows(user_prompts_data)
    
with open('openai_response.csv', 'w', newline='') as openai_response_file:
    csv_writer = csv.writer(openai_response_file)
    csv_writer.writerow(openai_response_fields)
    csv_writer.writerows(openai_response_data)
