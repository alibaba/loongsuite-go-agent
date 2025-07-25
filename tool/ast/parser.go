// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// AST Parser
// @@ N.B. DST framework provides a series of RestoreResolvers such
// as guess.New for resolving the package name from an importPath.
// However, its strategy is simply to guess by taking last section
// of the importpath as the package name. This can lead to issues
// where package names like github.com/foo/v2 are resolved as v2,
// while in reality, they might be foo. Incorrect resolutions can
// lead to some imports that should be present being rudely removed.
// To solve this issue, we disable DST's automatic Import management
// and use plain AST manipulation to add imports.

type AstParser struct {
	fset *token.FileSet
	dec  *decorator.Decorator
}

func NewAstParser() *AstParser {
	return &AstParser{
		fset: token.NewFileSet(),
	}
}

func (ap *AstParser) FindPosition(node dst.Node) token.Position {
	astNode := ap.dec.Ast.Nodes[node]
	if astNode == nil {
		return token.Position{Filename: "", Line: -1, Column: -1} // Invalid
	}
	return ap.fset.Position(astNode.Pos())
}

// ParseSnippet parses the AST from incomplete source code snippet.
func (ap *AstParser) ParseSnippet(codeSnippnet string) ([]dst.Stmt, error) {
	util.Assert(codeSnippnet != "", "empty code snippet")
	snippet := "package main; func _() {" + codeSnippnet + "}"
	file, err := decorator.ParseFile(ap.fset, "", snippet, 0)
	if err != nil {
		return nil, ex.Error(err)
	}
	return file.Decls[0].(*dst.FuncDecl).Body.List, nil
}

// ParseSource parses the AST from complete source code.
func (ap *AstParser) ParseSource(source string) (*dst.File, error) {
	util.Assert(source != "", "empty source")
	ap.dec = decorator.NewDecorator(ap.fset)
	dstRoot, err := ap.dec.Parse(source)
	if err != nil {
		return nil, ex.Error(err)
	}
	return dstRoot, nil
}

func (ap *AstParser) ParseFile(filePath string, mode parser.Mode) (*dst.File, error) {
	name := filepath.Base(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, ex.Error(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			ex.Fatal(err)
		}
	}(file)
	astFile, err := parser.ParseFile(ap.fset, name, file, mode)
	if err != nil {
		return nil, ex.Error(err)
	}
	ap.dec = decorator.NewDecorator(ap.fset)
	dstFile, err := ap.dec.DecorateFile(astFile)
	if err != nil {
		return nil, ex.Error(err)
	}
	return dstFile, nil
}

func ParseAstFromFileOnlyPackage(filePath string) (*dst.File, error) {
	return NewAstParser().ParseFile(filePath, parser.PackageClauseOnly)
}

func ParseAstFromFileFast(filePath string) (*dst.File, error) {
	return NewAstParser().ParseFile(filePath, parser.SkipObjectResolution)
}

// ParseAstFromFile parses the AST from complete source file.
func ParseAstFromFile(filePath string) (*dst.File, error) {
	return NewAstParser().ParseFile(filePath, parser.ParseComments)
}

// WriteAstToFile writes the AST to source file.
func WriteAstToFile(astRoot *dst.File, filePath string) (string, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return "", ex.Error(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			ex.Fatal(err)
		}
	}(file)

	r := decorator.NewRestorer()
	err = r.Fprint(file, astRoot)
	if err != nil {
		return "", ex.Error(err)
	}
	return file.Name(), nil
}
