import requests
import os
import pandas as pd
import time

def get_access_token():
    return os.environ.get('GITHUB_ACCESS_TOKEN')

def get_repository_info(repo_name):
    access_token = get_access_token()
    if not access_token:
        print("Access token not found. Please set GITHUB_ACCESS_TOKEN environment variable.")
        return
    headers = {'Authorization': f'Token {access_token}'}
    api_url = f'https://api.github.com/repos/{repo_name}'

    try:
        response = requests.get(api_url, headers=headers)
        time.sleep(0.5)

        if response.status_code == 200:
            repo_data = response.json()
            repo_info = {
                'RepoName': repo_data['name'],
                'Watch': repo_data['subscribers_count'],
                'Star': repo_data['stargazers_count'],
                'Fork': repo_data['forks_count']
            }
            return repo_info
        else:
            print(f"Fail to get {repo_name}:", response.status_code)
            print("Err info:", response.text)
    except requests.exceptions.RequestException as e:
        print(f"Request failed on {repo_name}:", e)

if __name__ == "__main__":
    intput_file = 'low.xlsx'
    df = pd.read_excel(intput_file)
    repo_info_list = []
    for repo_name in df['Repo']:
        repo_info = get_repository_info(repo_name)
        if repo_info:
            repo_info_list.append(repo_info)

    if repo_info_list:
        result_df = pd.DataFrame(repo_info_list)
        result_excel = 'repo_info_low.xlsx'
        result_df.to_excel(result_excel, index=False, engine='openpyxl')
        print(f"信息已保存至 {result_excel}")