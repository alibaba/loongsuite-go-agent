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
	"fmt"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/config"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
)

// == Code conjured by yyang, Jun. 2024 ==

// -----------------------------------------------------------------------------
// Trampoline Optimization
//
// Since trampoline-jump-if and trampoline functions are performance-critical,
// we are trying to optimize them as much as possible. The standard form of
// trampoline-jump-if looks like
//
//	if ctx, skip := otel_trampoline_onenter(&arg); skip {
//	    otel_trampoline_onexit(ctx, &retval)
//	    return ...
//	} else {
//	    defer otel_trampoline_onexit(ctx, &retval)
//	    ...
//	}
//
// The obvious optimization opportunities are cases when onEnter or onExit hooks
// are not present. For the latter case, we can replace the defer statement to
// empty statement, you might argue that we can remove the whole else block, but
// there might be more than one trampoline-jump-if in the same function, they are
// nested in the else block, i.e.
//
//	if ctx, skip := otel_trampoline_onenter(&arg); skip {
//	    otel_trampoline_onexit(ctx, &retval)
//	    return ...
//	} else {
//	    ;
//	    ...
//	}
//
// For the former case, it's a bit more complicated. We need to manually construct
// CallContext on the fly and pass it to onExit trampoline defer call and rewrite
// the whole condition to always false. The corresponding code snippet is
//
//	if false {
//	    ;
//	} else {
//	    defer otel_trampoline_onexit(&CallContext{...}, &retval)
//	    ...
//	}
//
// The if skeleton should be kept as is, otherwise inlining of trampoline-jump-if
// will not work. During compiling, the dce and sccp passes will remove the whole
// then block. That's not the whole story. We can further optimize the tjump iff
// the onEnter hook does not use SkipCall. In this case, we can rewrite condition
// of trampoline-jump-if to always false, remove return statement in then block,
// they are memory-aware and may generate memory SSA values during compilation.
//
//	if ctx,_ := otel_trampoline_onenter(&arg); false {
//	    ;
//	} else {
//	    defer otel_trampoline_onexit(ctx, &retval)
//	    ...
//	}
//
// The compiler responsible for hoisting the initialization statement out of the
// if skeleton, and the dce and sccp passes will remove the whole then block. All
// these trampoline functions looks as if they are executed sequentially, i.e.
//
//	ctx,_ := otel_trampoline_onenter(&arg);
//	defer otel_trampoline_onexit(ctx, &retval)
//
// Note that this optimization pass is fraigle as it really heavily depends on
// the structure of trampoline-jump-if and trampoline functions. Any change in
// tjump should be carefully examined.

// TJump describes a trampoline-jump-if optimization candidate
type TJump struct {
	target *dst.FuncDecl       // Target function we are hooking on
	ifStmt *dst.IfStmt         // Trampoline-jump-if statement
	rule   *rules.InstFuncRule // Rule associated with the trampoline-jump-if
}

func mustTJump(ifStmt *dst.IfStmt) {
	util.Assert(len(ifStmt.Decs.If) == 1, "must be a trampoline-jump-if")
	desc := ifStmt.Decs.If[0]
	util.Assert(desc == TJumpLabel, "must be a trampoline-jump-if")
}

func (rp *RuleProcessor) removeOnExitTrampolineCall(tjump *TJump) error {
	ifStmt := tjump.ifStmt
	elseBlock := ifStmt.Else.(*dst.BlockStmt)
	for i, stmt := range elseBlock.List {
		if _, ok := stmt.(*dst.DeferStmt); ok {
			// Replace defer statement with an empty statement
			elseBlock.List[i] = ast.EmptyStmt()
			if config.GetConf().Verbose {
				util.Log("Optimize tjump branch in %s",
					tjump.target.Name.Name)
			}
			break
		} else if _, ok := stmt.(*dst.IfStmt); ok {
			// Expected statement type and do nothing
		} else {
			// Unexpected statement type
			util.ShouldNotReachHereT("unexpected statement type")
		}
	}
	return nil
}

func replenishCallContextLiteral(tjump *TJump, expr dst.Expr) {
	rawFunc := tjump.target
	// Replenish call context literal with addresses of all arguments
	names := make([]dst.Expr, 0)
	for _, name := range getNames(rawFunc.Type.Params) {
		names = append(names, ast.AddressOf(ast.Ident(name)))
	}
	elems := expr.(*dst.UnaryExpr).X.(*dst.CompositeLit).Elts
	paramLiteral := elems[0].(*dst.KeyValueExpr).Value.(*dst.CompositeLit)
	paramLiteral.Elts = names
	// Replenish return values literal with addresses of all return values
	if rawFunc.Type.Results != nil {
		rets := make([]dst.Expr, 0)
		for _, name := range getNames(rawFunc.Type.Results) {
			rets = append(rets, ast.AddressOf(ast.Ident(name)))
		}
		elems = expr.(*dst.UnaryExpr).X.(*dst.CompositeLit).Elts
		returnLiteral := elems[1].(*dst.KeyValueExpr).Value.(*dst.CompositeLit)
		returnLiteral.Elts = rets
	}
}

// newCallContextImpl constructs a new CallContextImpl structure literal and
// replenishes its Params && ReturnValues field with addresses of all arguments.
// The CallContextImpl structure is used to pass arguments to the exit trampoline
func (rp *RuleProcessor) newCallContextImpl(tjump *TJump) (dst.Expr, error) {
	// TODO: This generated structure construction can also be marked via line
	// directive
	// One line please, otherwise debugging line number will be a nightmare
	tmpl := fmt.Sprintf("&CallContextImpl%s{Params:[]interface{}{},ReturnVals:[]interface{}{}}",
		rp.rule2Suffix[tjump.rule])
	p := ast.NewAstParser()
	astRoot, err := p.ParseSnippet(tmpl)
	if err != nil {
		return nil, ex.Error(err)
	}
	ctxExpr := astRoot[0].(*dst.ExprStmt).X
	// Replenish call context by passing addresses of all arguments
	replenishCallContextLiteral(tjump, ctxExpr)
	return ctxExpr, nil
}

func (rp *RuleProcessor) removeOnEnterTrampolineCall(tjump *TJump) error {
	// Construct CallContext on the fly and pass to onExit trampoline defer call
	callContextExpr, err := rp.newCallContextImpl(tjump)
	if err != nil {
		return ex.Error(err)
	}
	// Find defer call to onExit and replace its call context with new one
	found := false
	for _, stmt := range tjump.ifStmt.Else.(*dst.BlockStmt).List {
		// Replace call context argument of defer statement to structure literal
		if deferStmt, ok := stmt.(*dst.DeferStmt); ok {
			args := deferStmt.Call.Args
			util.Assert(len(args) >= 1, "must have at least one argument")
			args[0] = callContextExpr
			found = true
			break
		}
	}
	util.Assert(found, "defer statement not found")
	// Rewrite condition of trampoline-jump-if to always false and null out its
	// initialization statement and then block
	tjump.ifStmt.Init = nil
	tjump.ifStmt.Cond = ast.BoolFalse()
	tjump.ifStmt.Body = ast.Block(ast.EmptyStmt())
	if config.GetConf().Verbose {
		util.Log("Optimize tjump branch in %s", tjump.target.Name.Name)
	}
	// Remove generated onEnter trampoline function
	removed := rp.removeDeclWhen(func(d dst.Decl) bool {
		if funcDecl, ok := d.(*dst.FuncDecl); ok {
			return funcDecl.Name.Name == rp.makeName(tjump.rule, tjump.target, true)
		}
		return false
	})
	if removed == nil {
		return ex.Errorf(nil, "onEnter trampoline not found")
	}
	return nil
}

func flattenTJump(tjump *TJump, removedOnExit bool) {
	ifStmt := tjump.ifStmt
	initStmt := ifStmt.Init.(*dst.AssignStmt)
	util.Assert(len(initStmt.Lhs) == 2, "must be")

	ifStmt.Cond = ast.BoolFalse()
	ifStmt.Body = ast.Block(ast.EmptyStmt())

	if removedOnExit {
		// We removed the last reference to call context after nulling out body
		// block, at this point, all lhs are unused, replace assignment to simple
		// function call
		ifStmt.Init = ast.ExprStmt(initStmt.Rhs[0])
		// TODO: Remove onExit declaration
	} else {
		// Otherwise, mark skipCall identifier as unused
		skipCallIdent := initStmt.Lhs[1].(*dst.Ident)
		ast.MakeUnusedIdent(skipCallIdent)
	}
	if config.GetConf().Verbose {
		util.Log("Optimize skipCall in %s", tjump.target.Name.Name)
	}
}

func stripTJumpLabel(tjump *TJump) {
	ifStmt := tjump.ifStmt
	ifStmt.Decs.If = ifStmt.Decs.If[1:]
}

func (rp *RuleProcessor) optimizeTJumps() (err error) {
	for _, tjump := range rp.trampolineJumps {
		mustTJump(tjump.ifStmt)
		// Strip the trampoline-jump-if anchor label as no longer needed
		stripTJumpLabel(tjump)

		// No onExit hook present? Simply remove defer call to onExit trampoline.
		// Why we dont remove the whole else block of trampoline-jump-if? Well,
		// because there might be more than one trampoline-jump-if in the same
		// function, they are nested in the else block. See findJumpPoint for
		// more details.
		// TODO: Remove corresponding CallContextImpl methods
		rule := tjump.rule
		removedOnExit := false
		if rule.OnExit == "" {
			err = rp.removeOnExitTrampolineCall(tjump)
			if err != nil {
				return ex.Error(err)
			}
			removedOnExit = true
		}

		// No onEnter hook present? Construct CallContext on the fly and pass it
		// to onExit trampoline defer call and rewrite the whole condition to
		// always false, then null out its initialization statement.
		if rule.OnEnter == "" {
			err = rp.removeOnEnterTrampolineCall(tjump)
			if err != nil {
				return ex.Error(err)
			}
		}

		// No SkipCall used in onEnter hook? Rewrite cond of trampoline-jump-if
		// to always false, and remove return statement in then block, they are
		// memory aware and may generate memory SSA values during compilation.
		// This further simplifies the trampoline-jump-if and gives more chances
		// for optimization passes to kick in.
		if rule.OnEnter != "" {
			onEnterHook, err := getHookFunc(rule, true)
			if err != nil {
				return ex.Error(err)
			}
			foundPoison := false
			const poison = "SkipCall"
			// FIXME: We should traverse the call graph to find all possible
			// usage of SkipCall, but for now, we just check the onEnter hook
			// function body.
			dst.Inspect(onEnterHook, func(node dst.Node) bool {
				if ident, ok := node.(*dst.Ident); ok {
					if strings.Contains(ident.Name, poison) {
						foundPoison = true
						return false
					}
				}
				if foundPoison {
					return false
				}
				return true
			})
			if !foundPoison {
				flattenTJump(tjump, removedOnExit)
			}
		}
	}
	return nil
}
