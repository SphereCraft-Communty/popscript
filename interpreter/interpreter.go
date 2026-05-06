package interpreter

import (
	"fmt"
	"math/rand"

	"popscript/ast"
)

// Value wraps a runtime value with its type
type Value struct {
	Kind  string // "int", "float", "string", "bool"
	IVal  int64
	FVal  float64
	SVal  string
	BVal  bool
}

func (v Value) String() string {
	switch v.Kind {
	case "int":
		return fmt.Sprintf("%d", v.IVal)
	case "float":
		return fmt.Sprintf("%g", v.FVal)
	case "string":
		return v.SVal
	case "bool":
		if v.BVal {
			return "true"
		}
		return "false"
	}
	return "<nil>"
}

func intVal(n int64) Value   { return Value{Kind: "int", IVal: n} }
func floatVal(f float64) Value { return Value{Kind: "float", FVal: f} }
func strVal(s string) Value  { return Value{Kind: "string", SVal: s} }
func boolVal(b bool) Value   { return Value{Kind: "bool", BVal: b} }

// Environment holds variables
type Environment struct {
	vars map[string]Value
}

func newEnv() *Environment {
	return &Environment{vars: make(map[string]Value)}
}

func (e *Environment) set(name string, val Value) {
	e.vars[name] = val
}

func (e *Environment) get(name string) (Value, bool) {
	v, ok := e.vars[name]
	return v, ok
}

// Interpreter holds state
type Interpreter struct {
	env            *Environment
	importedModules map[string]bool   // module names
	importedSymbols map[string]string // symbol -> module
}

func New() *Interpreter {
	return &Interpreter{
		env:            newEnv(),
		importedModules: make(map[string]bool),
		importedSymbols: make(map[string]string),
	}
}

func (interp *Interpreter) Run(prog *ast.Program) error {
	for _, stmt := range prog.Statements {
		if err := interp.execStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (interp *Interpreter) execStatement(node ast.Node) error {
	switch n := node.(type) {
	case *ast.LibImport:
		return interp.execLibImport(n)
	case *ast.VarDecl:
		return interp.execVarDecl(n)
	case *ast.IfStmt:
		return interp.execIfStmt(n)
	case *ast.PrintStmt:
		return interp.execPrintStmt(n)
	case *ast.ExprStmt:
		_, err := interp.evalExpr(n.Expr)
		return err
	default:
		return fmt.Errorf("unknown statement type: %T", node)
	}
}

func (interp *Interpreter) execLibImport(n *ast.LibImport) error {
	interp.importedModules[n.Module] = true
	if n.Symbol != "" {
		interp.importedSymbols[n.Symbol] = n.Module
	}
	return nil
}

func (interp *Interpreter) execVarDecl(n *ast.VarDecl) error {
	val, err := interp.evalExpr(n.Value)
	if err != nil {
		return err
	}
	// Type coercion/check
	val, err = coerce(val, n.TypeName, n.Line)
	if err != nil {
		return err
	}
	interp.env.set(n.Name, val)
	return nil
}

func (interp *Interpreter) execIfStmt(n *ast.IfStmt) error {
	cond, err := interp.evalExpr(n.Condition)
	if err != nil {
		return err
	}
	if cond.Kind != "bool" {
		return fmt.Errorf("line %d: if condition must be bool, got %s", n.Line, cond.Kind)
	}
	if cond.BVal {
		for _, stmt := range n.Body {
			if err := interp.execStatement(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (interp *Interpreter) execPrintStmt(n *ast.PrintStmt) error {
	val, err := interp.evalExpr(n.Value)
	if err != nil {
		return err
	}
	fmt.Println(val.String())
	return nil
}

// ---- Expression evaluation ----

func (interp *Interpreter) evalExpr(node ast.Node) (Value, error) {
	switch n := node.(type) {
	case *ast.IntLit:
		return intVal(n.Value), nil
	case *ast.FloatLit:
		return floatVal(n.Value), nil
	case *ast.StringLit:
		return strVal(n.Value), nil
	case *ast.BoolLit:
		return boolVal(n.Value), nil
	case *ast.Identifier:
		v, ok := interp.env.get(n.Name)
		if !ok {
			return Value{}, fmt.Errorf("line %d: undefined variable %q", n.Line, n.Name)
		}
		return v, nil
	case *ast.BinaryExpr:
		return interp.evalBinary(n)
	case *ast.CallExpr:
		return interp.evalCall(n)
	default:
		return Value{}, fmt.Errorf("unknown expression type: %T", node)
	}
}

func (interp *Interpreter) evalBinary(n *ast.BinaryExpr) (Value, error) {
	left, err := interp.evalExpr(n.Left)
	if err != nil {
		return Value{}, err
	}
	right, err := interp.evalExpr(n.Right)
	if err != nil {
		return Value{}, err
	}

	// Promote int to float if mixed
	if left.Kind == "float" || right.Kind == "float" {
		lf := toFloat(left)
		rf := toFloat(right)
		switch n.Op {
		case "+":
			return floatVal(lf + rf), nil
		case "-":
			return floatVal(lf - rf), nil
		case "*":
			return floatVal(lf * rf), nil
		case "/":
			if rf == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return floatVal(lf / rf), nil
		case "<":
			return boolVal(lf < rf), nil
		case "<=":
			return boolVal(lf <= rf), nil
		case ">":
			return boolVal(lf > rf), nil
		case ">=":
			return boolVal(lf >= rf), nil
		case "==":
			return boolVal(lf == rf), nil
		case "!=":
			return boolVal(lf != rf), nil
		}
	}

	if left.Kind == "int" && right.Kind == "int" {
		switch n.Op {
		case "+":
			return intVal(left.IVal + right.IVal), nil
		case "-":
			return intVal(left.IVal - right.IVal), nil
		case "*":
			return intVal(left.IVal * right.IVal), nil
		case "/":
			if right.IVal == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return intVal(left.IVal / right.IVal), nil
		case "<":
			return boolVal(left.IVal < right.IVal), nil
		case "<=":
			return boolVal(left.IVal <= right.IVal), nil
		case ">":
			return boolVal(left.IVal > right.IVal), nil
		case ">=":
			return boolVal(left.IVal >= right.IVal), nil
		case "==":
			return boolVal(left.IVal == right.IVal), nil
		case "!=":
			return boolVal(left.IVal != right.IVal), nil
		}
	}

	if left.Kind == "string" && right.Kind == "string" {
		switch n.Op {
		case "+":
			return strVal(left.SVal + right.SVal), nil
		case "==":
			return boolVal(left.SVal == right.SVal), nil
		case "!=":
			return boolVal(left.SVal != right.SVal), nil
		}
	}

	if left.Kind == "bool" && right.Kind == "bool" {
		switch n.Op {
		case "==":
			return boolVal(left.BVal == right.BVal), nil
		case "!=":
			return boolVal(left.BVal != right.BVal), nil
		}
	}

	return Value{}, fmt.Errorf("unsupported operation %q on %s and %s", n.Op, left.Kind, right.Kind)
}

func (interp *Interpreter) evalCall(n *ast.CallExpr) (Value, error) {
	// Resolve module
	module := n.Module
	funcName := n.Func

	// If no module prefix, check if it's an imported symbol
	if module == "" {
		if mod, ok := interp.importedSymbols[funcName]; ok {
			module = mod
		}
	}

	// Dispatch to built-in stdlib
	switch module {
	case "random":
		return interp.callRandom(funcName, n.Args, n.Line)
	case "":
		// No module - check built-in functions
		return interp.callBuiltin(funcName, n.Args, n.Line)
	default:
		return Value{}, fmt.Errorf("line %d: unknown module %q", n.Line, module)
	}
}

func (interp *Interpreter) callRandom(fn string, args []ast.CallArg, line int) (Value, error) {
	switch fn {
	case "number":
		from, to, err := interp.resolveFromTo(args, line)
		if err != nil {
			return Value{}, err
		}
		result := from + rand.Int63n(to-from+1)
		return intVal(result), nil
	default:
		return Value{}, fmt.Errorf("line %d: unknown function random.%s", line, fn)
	}
}

func (interp *Interpreter) callBuiltin(fn string, args []ast.CallArg, line int) (Value, error) {
	return Value{}, fmt.Errorf("line %d: unknown function %q", line, fn)
}

func (interp *Interpreter) resolveFromTo(args []ast.CallArg, line int) (int64, int64, error) {
	named := make(map[string]Value)
	for _, arg := range args {
		val, err := interp.evalExpr(arg.Value)
		if err != nil {
			return 0, 0, err
		}
		if arg.Name != "" {
			named[arg.Name] = val
		}
	}
	fromV, ok1 := named["from"]
	toV, ok2 := named["to"]
	if !ok1 || !ok2 {
		return 0, 0, fmt.Errorf("line %d: random.number requires named args 'from' and 'to'", line)
	}
	from := toInt(fromV)
	to := toInt(toV)
	if from > to {
		return 0, 0, fmt.Errorf("line %d: 'from' must be <= 'to'", line)
	}
	return from, to, nil
}

// ---- Helpers ----

func toFloat(v Value) float64 {
	switch v.Kind {
	case "int":
		return float64(v.IVal)
	case "float":
		return v.FVal
	}
	return 0
}

func toInt(v Value) int64 {
	switch v.Kind {
	case "int":
		return v.IVal
	case "float":
		return int64(v.FVal)
	}
	return 0
}

func coerce(v Value, typeName string, line int) (Value, error) {
	switch typeName {
	case "int":
		switch v.Kind {
		case "int":
			return v, nil
		case "float":
			return intVal(int64(v.FVal)), nil
		}
	case "float":
		switch v.Kind {
		case "float":
			return v, nil
		case "int":
			return floatVal(float64(v.IVal)), nil
		}
	case "string":
		if v.Kind == "string" {
			return v, nil
		}
	case "bool":
		if v.Kind == "bool" {
			return v, nil
		}
	}
	return Value{}, fmt.Errorf("line %d: cannot assign %s to %s", line, v.Kind, typeName)
}
