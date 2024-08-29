# TXT file stores the results of the "gh search"
# this program de-duplicates and organizes these results into "Repo + Path" xlsx format

import pandas as pd

def process_links(input_file, output_file):
    with open(input_file, 'r', encoding='utf-16') as f:
        lines = f.readlines()

    data = []
    for line in lines:
        parts = line.strip().split(':')
        data.append({'仓库': parts[0], '路径': parts[1]})

    df = pd.DataFrame(data)
    df.drop_duplicates(inplace=True)
    df.to_excel(output_file, index=False)

if __name__ == "__main__":
    input_file = "high.txt"
    output_file = "high.xlsx"

    process_links(input_file, output_file)