# chaincode仓库

主要收集并整理Github仓库开源的Hyperledger Fabric链码文件，用于链码审计工具的测试及其他研究工作。

## Fabric链码

Hyperledger Fabric平台的智能合约最终会被打包成链码并部署在各个节点上。
Fabric平台的链码可以通过Go、Node.js、Java语言开发，Fabric为各个语言的开发者提供了相应的接口用于实现链码的各种逻辑功能。

## 收集方法

针对不同开发语言的必要库文件如下

- Go
  - (high level) github.com/hyperledger/fabric-contract-api-go/contractapi
  - (low level) github.com/hyperledger/fabric-chaincode-go/shim
- Node.js
  - (high level) fabric-contract-api
  - (low level) fabric-shim
- Java
  - org.hyperledger.fabric.contract

除了high level的库外之外，导入其他的库均需要实现以下两个接口：

- Init()
- Invoke()

因此，我们可以将库文件路径作为关键词进行搜索，然后选定语言，最后判断程序是否存在相关接口实现，如果存在，则该程序被认为是一个符合规范的链码。

## 注意事项

1. 由于链码开发大多存在跨文件的结构或函数复用，因此符合搜索要求的链码程序并不一定能够单独通过编译执行。除了下载链码文件外，有必要同时记录该链码的链接以供后续核验。
2. 下载过程中尽可能多记录一些属性，如链接，开发语言，行数，仓库热度（watch+fork+star），是否涉及PDC（链码中包含privatedata关键字），是否涉及跨链码调用（链码中包含invokechaincode关键字）
