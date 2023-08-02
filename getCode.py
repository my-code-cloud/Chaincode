import base64
import requests
import os
import pandas as pd
import time
from tqdm import tqdm

# get GitHub access tokrn
def get_access_token():
    return os.environ.get('GITHUB_ACCESS_TOKEN')

# get total line of code
def get_total_line(content):
    lines = content.splitlines()
    return len(lines)

# get repository info of given repo_name
def get_code(repo_name, code_path):
    access_token = get_access_token()
    if not access_token:
        print("Access token not found. Please set GITHUB_ACCESS_TOKEN environment variable.")
        return
    headers = {'Authorization': f'Token {access_token}'}
    api_url = f'https://api.github.com/repos/{repo_name}/contents/{code_path}'

    try:
        response = requests.get(api_url, headers=headers)
        time.sleep(0.2)

        if response.status_code == 200:
            content = response.json()
            code = base64.b64decode(content['content']).decode('utf-8')
            return code
        else:
            print(f"Fail to get {code_path}:", response.status_code)
            print("Err info:", response.text)
            return None
    except requests.exceptions.RequestException as e:
        print(f"Request failed on {repo_name}:", e)
        return None

if __name__ == "__main__":
    intput_file = 'go_chaincode_info_1.xlsx'
    df = pd.read_excel(intput_file)
    df['Lines'] = ''

    # init pbar
    total_codes = len(df['Repo'])
    with tqdm(total=total_codes, desc='get_code', unit='code') as pbar:
        for index, row in df.iterrows():
            repo_name = row['Repo']
            code_path = row['Path']
            code = get_code(repo_name, code_path)
            if code:
                with open(f'./go_chaincode_1/{index}.go', 'w', encoding='utf-8') as file:
                    file.write(code)
                total_line = get_total_line(code)
                df.at[index, 'Lines'] = total_line
            pbar.update(1)
    df.to_excel('go_chaincode_1.xlsx', index=False, engine='openpyxl')