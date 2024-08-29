def filter_unique_links(input_file, output_file):
    unique_links = set()

    with open(input_file, 'r') as infile:
        for line in infile:
            link = line.strip()
            if link not in unique_links:
                unique_links.add(link)

    with open(output_file, 'w') as outfile:
        for link in unique_links:
            outfile.write(link + '\n')

if __name__ == "__main__":
    input_file = 'total.txt'  # 输入文件名称
    output_file = 'unique_links.txt'  # 输出文件名称
    filter_unique_links(input_file, output_file)
    print("Unique links have been written to", output_file)
