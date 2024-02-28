import os
import pandas as pd

def check_keyword_in_file(file_path, keyword):
    """检查文件中是否包含特定关键字"""
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            return keyword in file.read()
    except:
        return False

def analyze_files_in_directory(directory_path, keyword):
    """分析目录中的文件并记录关键字出现情况"""
    data = {'File Name': [], 'Contains Keyword': []}
    for root, dirs, files in os.walk(directory_path):
        for file in files:
            file_path = os.path.join(root, file)
            file_name_without_extension = os.path.splitext(file)[0]
            contains_keyword = check_keyword_in_file(file_path, keyword)
            data['File Name'].append(file_name_without_extension)
            data['Contains Keyword'].append(1 if contains_keyword else 0)
    return pd.DataFrame(data)

# 使用函数
directory_path = './go_chaincode_1'  # 替换为您的目录路径
keyword = 'private'
result_df = analyze_files_in_directory(directory_path, keyword)

# 将结果保存为CSV文件
result_df.to_csv('file_analysis.csv', index=False)