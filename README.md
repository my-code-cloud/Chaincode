# Dataset

- This is the dataset for the paper "Understanding and Detecting Privacy Leakage Vulnerabilities in Hyperledger Fabric Chaincodes"
- It includes GitHub chaincodes and StackOverflow posts on privacy tasks based on Hyperledger Fabric blockchain, as well as the analysis and evaluation results for these data.

The directory structure is shown below:

```
.
├── LICENSE
├── README.md
├── evaluation
│   ├── code
│   ├── github_code_stats.xlsx
│   └── script
└── study
    ├── code
    ├── codeInfo.xlsx
    ├── script
    └── stackoverflowQA.xlsx
```

## Instruction

- **study** dictionary: Includes data for empirical studies on PDC misuse based on GitHub chaincodes in *codeInfo.xlsx* and StackOverflow posts in *stackoverflowQA.xlsx*. The *code* directory and the *script* directory store the source code and the scripts used to process the code, respectively.
- **evaluation** dictionary: Includes test data used to validate our tool in *code* directory and the results in *github_code_stats.xlsx*. the *script* directory store the scripts used to process the code, too.

## Reproducibility Instructions

1. Make sure you have PDChecker installed and tested. Refer to [PDChecker](https://github.com/zm-stack/PDChecker)

```bash
revive -h
```

2. Get into the following directory. This directory includes all the test cases evaluating PDChecker in the paper.

```bash
cd evaluation/code directory
```

3. Call PDCherker to test all codes and save the results to the *output* file. If successful, you will get the same results as in the *result* file.

```bash
revive -config chaincode.toml -formatter stylish > output
```

1. Check all possible vulnerability in each code and reconcile each vulnerability report. You can refer to *github_code_stats.xlsx* for the final results. It records the vulnerabilities found in each code by manual check, as well as all false positives and false negatives found in the vulnerability reports.

Here, the reproduction step ends and you can compare the results to TABLE IV in the paper to evaluate reproducibility.
