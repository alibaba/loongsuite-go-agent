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

package instrument

import (
	_ "embed"
	"go/token"
	"strconv"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

// -----------------------------------------------------------------------------
// Trampoline Jump
//
// We distinguish between three types of functions: RawFunc, TrampolineFunc, and
// HookFunc. RawFunc is the original function that needs to be instrumented.
// TrampolineFunc is the function that is generated to call the onEnter and
// onExit hooks, it serves as a trampoline to the original function. HookFunc is
// the function that is called at entrypoint and exitpoint of the RawFunc. The
// so-called "Trampoline Jump" snippet is inserted at start of raw func, it is
// guaranteed to be generated within one line to avoid confusing debugging, as
// its name suggests, it jumps to the trampoline function from raw function.
const (
	TrampolineSetParamName           = "SetParam"
	TrampolineGetParamName           = "GetParam"
	TrampolineSetReturnValName       = "SetReturnVal"
	TrampolineGetReturnValName       = "GetReturnVal"
	TrampolineValIdentifier          = "val"
	TrampolineCtxIdentifier          = "c"
	TrampolineParamsIdentifier       = "Params"
	TrampolineFuncNameIdentifier     = "FuncName"
	TrampolinePackageNameIdentifier  = "PackageName"
	TrampolineReturnValsIdentifier   = "ReturnVals"
	TrampolineSkipName               = "skip"
	TrampolineCallContextName        = "callContext"
	TrampolineCallContextType        = "CallContext"
	TrampolineCallContextImplType    = "CallContextImpl"
	TrampolineOnEnterName            = "OtelOnEnterTrampoline"
	TrampolineOnExitName             = "OtelOnExitTrampoline"
	TrampolineOnEnterNamePlaceholder = "\"OtelOnEnterNamePlaceholder\""
	TrampolineOnExitNamePlaceholder  = "\"OtelOnExitNamePlaceholder\""
)

// @@ Modification on this trampoline template should be cautious, as it imposes
// many implicit constraints on generated code, known constraints are as follows:
// - It's performance critical, so it should be as simple as possible
// - It should not import any package because there is no guarantee that package
//   is existed in import config during the compilation, one practical approach
//   is to use function variables and setup these variables in preprocess stage
// - It should not panic as this affects user application
// - Function and variable names are coupled with the framework, any modification
//   on them should be synced with the framework

//go:embed template.go
var trampolineTemplate string

func (rp *RuleProcessor) materializeTemplate() error {
	// Read trampoline template and materialize onEnter and onExit function
	// declarations based on that
	p := ast.NewAstParser()
	astRoot, err := p.ParseSource(trampolineTemplate)
	if err != nil {
		return ex.Error(err)
	}

	rp.varDecls = make([]dst.Decl, 0)
	rp.callCtxMethods = make([]*dst.FuncDecl, 0)
	for _, node := range astRoot.Decls {
		// Materialize function declarations
		if decl, ok := node.(*dst.FuncDecl); ok {
			if decl.Name.Name == TrampolineOnEnterName {
				rp.onEnterHookFunc = decl
				rp.addDecl(decl)
			} else if decl.Name.Name == TrampolineOnExitName {
				rp.onExitHookFunc = decl
				rp.addDecl(decl)
			} else if ast.HasReceiver(decl) {
				// We know exactly this is CallContextImpl method
				t := decl.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident).Name
				util.Assert(t == TrampolineCallContextImplType, "sanity check")
				rp.callCtxMethods = append(rp.callCtxMethods, decl)
				rp.addDecl(decl)
			}
		}
		// Materialize variable declarations
		if decl, ok := node.(*dst.GenDecl); ok {
			// No further processing for variable declarations, just append them
			if decl.Tok == token.VAR {
				rp.varDecls = append(rp.varDecls, decl)
			} else if decl.Tok == token.TYPE {
				rp.callCtxDecl = decl
				rp.addDecl(decl)
			}
		}
	}
	util.Assert(rp.callCtxDecl != nil, "sanity check")
	util.Assert(len(rp.varDecls) > 0, "sanity check")
	util.Assert(rp.onEnterHookFunc != nil, "sanity check")
	util.Assert(rp.onExitHookFunc != nil, "sanity check")
	return nil
}

func getNames(list *dst.FieldList) []string {
	var names []string
	for _, field := range list.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

func makeOnXName(t *rules.InstFuncRule, onEnter bool) string {
	if onEnter {
		return t.OnEnter
	} else {
		return t.OnExit
	}
}

type ParamTrait struct {
	Index          int
	IsVaradic      bool
	IsInterfaceAny bool
}

func isHookDefined(root *dst.File, rule *rules.InstFuncRule) bool {
	util.Assert(rule.OnEnter != "" || rule.OnExit != "", "hook must be set")
	if rule.OnEnter != "" {
		if ast.FindFuncDecl(root, rule.OnEnter) == nil {
			return false
		}
	}
	if rule.OnExit != "" {
		if ast.FindFuncDecl(root, rule.OnExit) == nil {
			return false
		}
	}
	return true
}

func findHookFile(rule *rules.InstFuncRule) (string, error) {
	files, err := findRuleFiles(rule)
	if err != nil {
		return "", ex.Error(err)
	}
	for _, file := range files {
		if !util.IsGoFile(file) {
			continue
		}
		root, err := ast.ParseAstFromFileFast(file)
		if err != nil {
			return "", ex.Error(err)
		}
		if isHookDefined(root, rule) {
			return file, nil
		}
	}
	return "", ex.Errorf(nil, "no hook %s/%s found for %s from %v",
		rule.OnEnter, rule.OnExit, rule.Function, files)
}

func findRuleFiles(rule rules.InstRule) ([]string, error) {
	files, err := util.ListFiles(rule.GetPath())
	if err != nil {
		return nil, ex.Error(err)
	}
	switch rule.(type) {
	case *rules.InstFuncRule, *rules.InstFileRule:
		return files, nil
	case *rules.InstStructRule:
		util.ShouldNotReachHereT("insane rule type")
	}
	return nil, nil
}

func getHookFunc(t *rules.InstFuncRule, onEnter bool) (*dst.FuncDecl, error) {
	file, err := findHookFile(t)
	if err != nil {
		return nil, ex.Error(err)
	}
	astRoot, err := ast.ParseAstFromFile(file)
	if err != nil {
		return nil, ex.Error(err)
	}
	var target *dst.FuncDecl
	if onEnter {
		target = ast.FindFuncDecl(astRoot, t.OnEnter)
	} else {
		target = ast.FindFuncDecl(astRoot, t.OnExit)
	}
	if target != nil {
		return target, nil
	}

	if onEnter {
		return nil, ex.Errorf(err, "hook %s", t.OnEnter)
	} else {
		return nil, ex.Errorf(err, "hook %s", t.OnExit)
	}
}

func getHookParamTraits(t *rules.InstFuncRule, onEnter bool) ([]ParamTrait, error) {
	target, err := getHookFunc(t, onEnter)
	if err != nil {
		return nil, ex.Error(err)
	}
	var attrs []ParamTrait
	// Find which parameter is type of interface{}
	for i, field := range target.Type.Params.List {
		attr := ParamTrait{Index: i}
		if ast.IsInterfaceType(field.Type) {
			attr.IsInterfaceAny = true
		}
		if ast.IsEllipsis(field.Type) {
			attr.IsVaradic = true
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}

func (rp *RuleProcessor) callOnEnterHook(t *rules.InstFuncRule, traits []ParamTrait) error {
	// The actual parameter list of hook function should be the same as the
	// target function
	if rp.exact {
		util.Assert(len(traits) == (len(rp.onEnterHookFunc.Type.Params.List)+1),
			"do you miss api.CallContext parameter?")
	}
	// Hook: 	   func onEnterFoo(callContext* CallContext, p*[]int)
	// Trampoline: func OtelOnEnterTrampoline_foo(p *[]int)
	args := []dst.Expr{dst.NewIdent(TrampolineCallContextName)}
	if rp.exact {
		for idx, field := range rp.onEnterHookFunc.Type.Params.List {
			trait := traits[idx+1 /*CallContext*/]
			for _, name := range field.Names { // syntax of n1,n2 type
				if trait.IsVaradic {
					args = append(args, ast.DereferenceOf(ast.Ident(name.Name+"...")))
				} else {
					args = append(args, ast.DereferenceOf(dst.NewIdent(name.Name)))
				}
			}
		}
	}
	fnName := makeOnXName(t, true)
	call := ast.ExprStmt(ast.CallTo(fnName, args))
	iff := ast.IfNotNilStmt(
		dst.NewIdent(fnName),
		ast.Block(call),
		nil,
	)
	insertAt(rp.onEnterHookFunc, iff, len(rp.onEnterHookFunc.Body.List)-1)
	return nil
}

func (rp *RuleProcessor) callOnExitHook(t *rules.InstFuncRule, traits []ParamTrait) error {
	// The actual parameter list of hook function should be the same as the
	// target function
	if rp.exact {
		util.Assert(len(traits) == len(rp.onExitHookFunc.Type.Params.List),
			"do you miss api.CallContext parameter?")
	}
	// Hook: 	   func onExitFoo(ctx* CallContext, p*[]int)
	// Trampoline: func OtelOnExitTrampoline_foo(ctx* CallContext, p *[]int)
	var args []dst.Expr
	for idx, field := range rp.onExitHookFunc.Type.Params.List {
		if idx == 0 {
			args = append(args, dst.NewIdent(TrampolineCallContextName))
			if !rp.exact {
				// Generic hook function, no need to process parameters
				break
			}
			continue
		}
		trait := traits[idx]
		for _, name := range field.Names { // syntax of n1,n2 type
			if trait.IsVaradic {
				arg := ast.DereferenceOf(ast.Ident(name.Name + "..."))
				args = append(args, arg)
			} else {
				arg := ast.DereferenceOf(dst.NewIdent(name.Name))
				args = append(args, arg)
			}
		}
	}
	fnName := makeOnXName(t, false)
	call := ast.ExprStmt(ast.CallTo(fnName, args))
	iff := ast.IfNotNilStmt(
		dst.NewIdent(fnName),
		ast.Block(call),
		nil,
	)
	insertAtEnd(rp.onExitHookFunc, iff)
	return nil
}

func rectifyAnyType(paramList *dst.FieldList, traits []ParamTrait) error {
	if len(paramList.List) != len(traits) {
		return ex.Errorf(nil, "do you miss api.CallContext parameter?")
	}
	for i, field := range paramList.List {
		trait := traits[i]
		if trait.IsInterfaceAny {
			// Rectify type to "interface{}"
			field.Type = ast.InterfaceType()
		}
	}
	return nil
}

func (rp *RuleProcessor) addHookFuncVar(t *rules.InstFuncRule,
	traits []ParamTrait, onEnter bool) error {
	paramTypes := &dst.FieldList{List: []*dst.Field{}}
	if rp.exact {
		paramTypes = rp.buildTrampolineType(onEnter)
	}
	addCallContext(paramTypes)
	if rp.exact {
		// Hook functions may uses interface{} as parameter type, as some types of
		// raw function is not exposed
		err := rectifyAnyType(paramTypes, traits)
		if err != nil {
			return ex.Error(err)
		}
	}

	// Generate var decl and append it to the target file, note that many target
	// functions may match the same hook function, it's a fatal error to append
	// multiple hook function declarations to the same file, so we need to check
	// if the hook function variable is already declared in the target file
	exist := false
	fnName := makeOnXName(t, onEnter)
	funcDecl := &dst.FuncDecl{
		Name: &dst.Ident{
			Name: fnName,
		},
		Type: &dst.FuncType{
			Func:   false,
			Params: paramTypes,
		},
	}
	for _, decl := range rp.target.Decls {
		if fDecl, ok := decl.(*dst.FuncDecl); ok {
			if fDecl.Name.Name == fnName {
				exist = true
				break
			}
		}
	}
	if !exist {
		rp.addDecl(funcDecl)
	}
	return nil
}

func insertAt(funcDecl *dst.FuncDecl, stmt dst.Stmt, index int) {
	stmts := funcDecl.Body.List
	newStmts := append(stmts[:index],
		append([]dst.Stmt{stmt}, stmts[index:]...)...)
	funcDecl.Body.List = newStmts
}

func insertAtEnd(funcDecl *dst.FuncDecl, stmt dst.Stmt) {
	insertAt(funcDecl, stmt, len(funcDecl.Body.List))
}

func (rp *RuleProcessor) renameFunc(t *rules.InstFuncRule) {
	// Randomize trampoline function names
	rp.onEnterHookFunc.Name.Name = rp.makeName(t, rp.rawFunc, true)
	dst.Inspect(rp.onEnterHookFunc, func(node dst.Node) bool {
		if basicLit, ok := node.(*dst.BasicLit); ok {
			// Replace OtelOnEnterTrampolinePlaceHolder to real hook func name
			if basicLit.Value == TrampolineOnEnterNamePlaceholder {
				basicLit.Value = strconv.Quote(t.OnEnter)
			}
		}
		return true
	})
	rp.onExitHookFunc.Name.Name = rp.makeName(t, rp.rawFunc, false)
	dst.Inspect(rp.onExitHookFunc, func(node dst.Node) bool {
		if basicLit, ok := node.(*dst.BasicLit); ok {
			if basicLit.Value == TrampolineOnExitNamePlaceholder {
				basicLit.Value = strconv.Quote(t.OnExit)
			}
		}
		return true
	})
}

func addCallContext(list *dst.FieldList) {
	callCtx := ast.NewField(
		TrampolineCallContextName,
		dst.NewIdent(TrampolineCallContextType),
	)
	list.List = append([]*dst.Field{callCtx}, list.List...)
}

func (rp *RuleProcessor) buildTrampolineType(onEnter bool) *dst.FieldList {
	paramList := &dst.FieldList{List: []*dst.Field{}}
	if onEnter {
		if ast.HasReceiver(rp.rawFunc) {
			recvField := dst.Clone(rp.rawFunc.Recv.List[0]).(*dst.Field)
			paramList.List = append(paramList.List, recvField)
		}
		for _, field := range rp.rawFunc.Type.Params.List {
			paramField := dst.Clone(field).(*dst.Field)
			paramList.List = append(paramList.List, paramField)
		}
	} else {
		if rp.rawFunc.Type.Results != nil {
			for _, field := range rp.rawFunc.Type.Results.List {
				retField := dst.Clone(field).(*dst.Field)
				paramList.List = append(paramList.List, retField)
			}
		}
	}
	return paramList
}

func (rp *RuleProcessor) rectifyTypes() {
	onEnterHookFunc, onExitHookFunc := rp.onEnterHookFunc, rp.onExitHookFunc
	onEnterHookFunc.Type.Params = rp.buildTrampolineType(true)
	onExitHookFunc.Type.Params = rp.buildTrampolineType(false)
	candidate := []*dst.FieldList{
		onEnterHookFunc.Type.Params,
		onExitHookFunc.Type.Params,
	}
	for _, list := range candidate {
		for i := 0; i < len(list.List); i++ {
			paramField := list.List[i]
			paramFieldType := desugarType(paramField)
			paramField.Type = ast.DereferenceOf(paramFieldType)
		}
	}
	addCallContext(onExitHookFunc.Type.Params)
}

// replenishCallContext replenishes the call context before hook invocation
func (rp *RuleProcessor) replenishCallContext(onEnter bool) bool {
	funcDecl := rp.onEnterHookFunc
	if !onEnter {
		funcDecl = rp.onExitHookFunc
	}
	for _, stmt := range funcDecl.Body.List {
		if assignStmt, ok := stmt.(*dst.AssignStmt); ok {
			lhs := assignStmt.Lhs
			if sel, ok := lhs[0].(*dst.SelectorExpr); ok {
				switch sel.Sel.Name {
				case TrampolineFuncNameIdentifier:
					util.Assert(onEnter, "sanity check")
					// callContext.FuncName = "..."
					rhs := assignStmt.Rhs
					if len(rhs) == 1 {
						rhsExpr := rhs[0]
						if basicLit, ok := rhsExpr.(*dst.BasicLit); ok {
							if basicLit.Kind == token.STRING {
								rawFuncName := rp.rawFunc.Name.Name
								basicLit.Value = strconv.Quote(rawFuncName)
							} else {
								return false // ill-formed AST
							}
						} else {
							return false // ill-formed AST
						}
					} else {
						return false // ill-formed AST
					}
				case TrampolinePackageNameIdentifier:
					util.Assert(onEnter, "sanity check")
					// callContext.PackageName = "..."
					rhs := assignStmt.Rhs
					if len(rhs) == 1 {
						rhsExpr := rhs[0]
						if basicLit, ok := rhsExpr.(*dst.BasicLit); ok {
							if basicLit.Kind == token.STRING {
								pkgName := rp.target.Name.Name
								basicLit.Value = strconv.Quote(pkgName)
							} else {
								return false // ill-formed AST
							}
						} else {
							return false // ill-formed AST
						}
					} else {
						return false // ill-formed AST
					}
				default:
					// callContext.Params = []interface{}{...} or
					// callContext.(*CallContextImpl).Params[0] = &int
					rhs := assignStmt.Rhs
					if len(rhs) == 1 {
						rhsExpr := rhs[0]
						if compositeLit, ok := rhsExpr.(*dst.CompositeLit); ok {
							elems := compositeLit.Elts
							names := getNames(funcDecl.Type.Params)
							for i, name := range names {
								if i == 0 && !onEnter {
									// SKip first callContext parameter for onExit
									continue
								}
								elems = append(elems, ast.Ident(name))
							}
							compositeLit.Elts = elems
						} else {
							return false // ill-formed AST
						}
					} else {
						return false // ill-formed AST
					}
				}
			}

		}
	}
	return true
}

// -----------------------------------------------------------------------------
// Dynamic CallContext API Generation
//
// This is somewhat challenging, as we need to generate type-aware CallContext
// APIs, which means we need to generate a bunch of switch statements to handle
// different types of parameters. Different RawFuncs in the same package may have
// different types of parameters, all of them should have their own CallContext
// implementation, thus we need to generate a bunch of CallContextImpl{suffix}
// types and methods to handle them. The suffix is generated based on the rule
// suffix, so that we can distinguish them from each other.

// implementCallContext effectively "implements" the CallContext interface by
// renaming occurrences of CallContextImpl to CallContextImpl{suffix} in the
// trampoline template
func (rp *RuleProcessor) implementCallContext(t *rules.InstFuncRule) {
	suffix := rp.rule2Suffix[t]
	structType := rp.callCtxDecl.Specs[0].(*dst.TypeSpec)
	util.Assert(structType.Name.Name == TrampolineCallContextImplType,
		"sanity check")
	structType.Name.Name += suffix             // type declaration
	for _, method := range rp.callCtxMethods { // method declaration
		method.Recv.List[0].Type.(*dst.StarExpr).X.(*dst.Ident).Name += suffix
	}
	for _, node := range []dst.Node{rp.onEnterHookFunc, rp.onExitHookFunc} {
		dst.Inspect(node, func(node dst.Node) bool {
			if ident, ok := node.(*dst.Ident); ok {
				if ident.Name == TrampolineCallContextImplType {
					ident.Name += suffix
					return false
				}
			}
			return true
		})
	}
}

func setValue(field string, idx int, typ dst.Expr) *dst.CaseClause {
	// *(c.Params[idx].(*int)) = val.(int)
	// c.Params[idx] = val iff type is interface{}
	se := ast.SelectorExpr(ast.Ident(TrampolineCtxIdentifier), field)
	ie := ast.IndexExpr(se, ast.IntLit(idx))
	te := ast.TypeAssertExpr(ie, ast.DereferenceOf(typ))
	pe := ast.ParenExpr(te)
	de := ast.DereferenceOf(pe)
	val := ast.Ident(TrampolineValIdentifier)
	assign := ast.AssignStmt(de, ast.TypeAssertExpr(val, typ))
	if ast.IsInterfaceType(typ) {
		assign = ast.AssignStmt(ie, val)
	}
	caseClause := ast.SwitchCase(
		ast.Exprs(ast.IntLit(idx)),
		ast.Stmts(assign),
	)
	return caseClause
}

func getValue(field string, idx int, typ dst.Expr) *dst.CaseClause {
	// return *(c.Params[idx].(*int))
	// return c.Params[idx] iff type is interface{}
	se := ast.SelectorExpr(ast.Ident(TrampolineCtxIdentifier), field)
	ie := ast.IndexExpr(se, ast.IntLit(idx))
	te := ast.TypeAssertExpr(ie, ast.DereferenceOf(typ))
	pe := ast.ParenExpr(te)
	de := ast.DereferenceOf(pe)
	ret := ast.ReturnStmt(ast.Exprs(de))
	if ast.IsInterfaceType(typ) {
		ret = ast.ReturnStmt(ast.Exprs(ie))
	}
	caseClause := ast.SwitchCase(
		ast.Exprs(ast.IntLit(idx)),
		ast.Stmts(ret),
	)
	return caseClause
}

func getParamClause(idx int, typ dst.Expr) *dst.CaseClause {
	return getValue(TrampolineParamsIdentifier, idx, typ)
}

func setParamClause(idx int, typ dst.Expr) *dst.CaseClause {
	return setValue(TrampolineParamsIdentifier, idx, typ)
}

func getReturnValClause(idx int, typ dst.Expr) *dst.CaseClause {
	return getValue(TrampolineReturnValsIdentifier, idx, typ)
}

func setReturnValClause(idx int, typ dst.Expr) *dst.CaseClause {
	return setValue(TrampolineReturnValsIdentifier, idx, typ)
}

// desugarType desugars parameter type to its original type, if parameter
// is type of ...T, it will be converted to []T
func desugarType(param *dst.Field) dst.Expr {
	if ft, ok := param.Type.(*dst.Ellipsis); ok {
		return ast.ArrayType(ft.Elt)
	}
	return param.Type
}

func (rp *RuleProcessor) rewriteCallContextImpl() {
	util.Assert(len(rp.callCtxMethods) > 4, "sanity check")
	var (
		methodSetParam  *dst.FuncDecl
		methodGetParam  *dst.FuncDecl
		methodGetRetVal *dst.FuncDecl
		methodSetRetVal *dst.FuncDecl
	)
	for _, decl := range rp.callCtxMethods {
		switch decl.Name.Name {
		case TrampolineSetParamName:
			methodSetParam = decl
		case TrampolineGetParamName:
			methodGetParam = decl
		case TrampolineGetReturnValName:
			methodGetRetVal = decl
		case TrampolineSetReturnValName:
			methodSetRetVal = decl
		}
	}
	// Rewrite SetParam and GetParam methods
	// Dont believe what you see in template.go, we will null out it and rewrite
	// the whole switch statement
	methodSetParamBody := methodSetParam.Body.List[1].(*dst.SwitchStmt).Body
	methodGetParamBody := methodGetParam.Body.List[0].(*dst.SwitchStmt).Body
	methodSetRetValBody := methodSetRetVal.Body.List[1].(*dst.SwitchStmt).Body
	methodGetRetValBody := methodGetRetVal.Body.List[0].(*dst.SwitchStmt).Body
	methodGetParamBody.List = nil
	methodSetParamBody.List = nil
	methodGetRetValBody.List = nil
	methodSetRetValBody.List = nil
	idx := 0
	if ast.HasReceiver(rp.rawFunc) {
		recvType := rp.rawFunc.Recv.List[0].Type
		clause := setParamClause(idx, recvType)
		methodSetParamBody.List = append(methodSetParamBody.List, clause)
		clause = getParamClause(idx, recvType)
		methodGetParamBody.List = append(methodGetParamBody.List, clause)
		idx++
	}
	for _, param := range rp.rawFunc.Type.Params.List {
		paramType := desugarType(param)
		for range param.Names {
			clause := setParamClause(idx, paramType)
			methodSetParamBody.List =
				append(methodSetParamBody.List, clause)
			clause = getParamClause(idx, paramType)
			methodGetParamBody.List =
				append(methodGetParamBody.List, clause)
			idx++
		}
	}
	// Rewrite GetReturnVal and SetReturnVal methods
	if rp.rawFunc.Type.Results != nil {
		idx = 0
		for _, retval := range rp.rawFunc.Type.Results.List {
			retType := desugarType(retval)
			for range retval.Names {
				clause := getReturnValClause(idx, retType)
				methodGetRetValBody.List =
					append(methodGetRetValBody.List, clause)
				clause = setReturnValClause(idx, retType)
				methodSetRetValBody.List =
					append(methodSetRetValBody.List, clause)
				idx++
			}
		}
	}
}

func (rp *RuleProcessor) callHookFunc(t *rules.InstFuncRule,
	onEnter bool) error {
	traits, err := getHookParamTraits(t, onEnter)
	if err != nil {
		return ex.Error(err)
	}
	err = rp.addHookFuncVar(t, traits, onEnter)
	if err != nil {
		return ex.Error(err)
	}
	if onEnter {
		err = rp.callOnEnterHook(t, traits)
	} else {
		err = rp.callOnExitHook(t, traits)
	}
	if err != nil {
		return ex.Error(err)
	}
	if !rp.replenishCallContext(onEnter) {
		return ex.Errorf(nil, "can not rewrite hook function")
	}
	return nil
}

func (rp *RuleProcessor) generateTrampoline(t *rules.InstFuncRule) error {
	// Materialize various declarations from template file, no one wants to see
	// a bunch of manual AST code generation, isn't it?
	err := rp.materializeTemplate()
	if err != nil {
		return ex.Error(err)
	}
	// Implement CallContext interface
	rp.implementCallContext(t)
	// Rewrite type-aware CallContext APIs
	rp.rewriteCallContextImpl()
	// Rename trampoline functions
	rp.renameFunc(t)
	// Rectify types of trampoline functions
	rp.rectifyTypes()
	// Generate calls to hook functions
	if t.OnEnter != "" {
		err = rp.callHookFunc(t, true)
		if err != nil {
			return ex.Error(err)
		}
	}
	if t.OnExit != "" {
		err = rp.callHookFunc(t, false)
		if err != nil {
			return ex.Error(err)
		}
	}
	return nil
}
