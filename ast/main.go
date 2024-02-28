package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	// 打开源代码文件
	sourceCode, err := os.Open("./private.go")
	if err != nil {
		fmt.Printf("无法打开文件：%s\n", err)
		return
	}
	defer sourceCode.Close()

	// 创建文件集
	fset := token.NewFileSet()

	// 解析源代码文件
	node, err := parser.ParseFile(fset, "source.go", sourceCode, parser.AllErrors)
	if err != nil {
		fmt.Printf("解析源代码文件时发生错误：%s\n", err)
		return
	}

	// 输出AST结构
	fmt.Println("抽象语法树 (AST) 结构：")
	ast.Print(fset, node)
}
