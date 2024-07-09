def filter_links(a_file, b_file, output_file):
    # 读取文件B中的所有链接到一个集合中
    b_links = set()
    with open(b_file, 'r') as b:
        for line in b:
            b_links.add(line.strip())
    
    # 读取文件A，过滤掉存在于B中的链接，并写入输出文件
    with open(a_file, 'r') as a, open(output_file, 'w') as output:
        for line in a:
            link = line.strip()
            if link not in b_links:
                output.write(link + '\n')

if __name__ == "__main__":
    a_file = 'total_unique.txt'  # 文件A的名称
    b_file = 'former_links.txt'  # 文件B的名称
    output_file = 'unique_links_2.txt'  # 输出文件的名称
    filter_links(a_file, b_file, output_file)
    print("Filtered links have been written to", output_file)
