# 处理md文件中从github上获取到的链码信息，得到“仓库+路径”的格式

import pandas as pd

def process_links(input_file, output_file):
    # 读取文本文件并按行划分内容
    with open(input_file, 'r', encoding='utf-16') as f:
        lines = f.readlines()

    # 将内容按:分隔，并创建一个字典列表
    data = []
    for line in lines:
        parts = line.strip().split(':')
        data.append({'仓库': parts[0], '路径': parts[1]})

    # 将字典列表转换为DataFrame
    df = pd.DataFrame(data)

    # 过滤掉重复数据
    df.drop_duplicates(inplace=True)

    # 将数据保存到Excel文件
    df.to_excel(output_file, index=False)

if __name__ == "__main__":
    # 输入和输出文件名
    input_file = "high.txt"
    output_file = "high.xlsx"

    process_links(input_file, output_file)