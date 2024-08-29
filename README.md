# Dataset

- This is the dataset for the paper "Understanding and Detecting Privacy Leakage Vulnerabilities in Hyperledger Fabric Chaincodes"
- It includes GitHub chaincodes and StackOverflow posts on privacy tasks based on Hyperledger Fabric blockchain, as well as the analysis and evaluation results for these data.

The directory structure is shown below:

  Dataset
  ├── README.md
  ├── study
  │    ├── code
  │    ├── script
  │    ├── codeInfo.xlsx
  │    └── stackoverflowQA.xlsx
  └── evaluation
        ├── code
        ├── script
        └── github_code_stats.xlsx

---

## Instruction

- **study** dictionary: Includes data for empirical studies on PDC misuse based on GitHub chaincodes in *codeInfo.xlsx* and StackOverflow posts in *stackoverflowQA.xlsx*. The *code* directory and the *script* directory store the source code and the scripts used to process the code, respectively.
- **evaluation** dictionary: Includes test data used to validate our tool in *code* directory and the results in *github_code_stats.xlsx*. the *script* directory store the scripts used to process the code, too.

## Reproducibility Instructions:

```bash
# Make sure you have PDChecker installed and tested
# Run the following command in evaluation/code directory
revive -config chaincode.toml -formatter stylish > output
```

You will see all the test results in the output file.
