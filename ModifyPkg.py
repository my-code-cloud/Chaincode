import os
import re

def modify_package_to_main(file_path):
    """修改Go文件的package声明为main"""
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            content = file.read()

        # 使用正则表达式查找并替换package声明
        modified_content = re.sub(r'^package\s+\w+', 'package main', content, flags=re.MULTILINE)

        with open(file_path, 'w', encoding='utf-8') as file:
            file.write(modified_content)
    except Exception as e:
        print(f"Error processing {file_path}: {e}")

def modify_packages_in_directory(directory_path):
    """遍历目录并修改所有Go文件的package声明"""
    for root, dirs, files in os.walk(directory_path):
        for file in files:
            if file.endswith('.go'):
                file_path = os.path.join(root, file)
                modify_package_to_main(file_path)

# 使用函数
directory_path = './go_chaincode_1'  # 替换为您的目录路径
modify_packages_in_directory(directory_path)