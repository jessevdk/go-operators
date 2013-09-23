package types

import (
	"go/ast"
	"go/token"
	"strings"
)

type OverloadInfo struct {
	Func *Func
	Recv ast.Expr
	Oper ast.Expr
}

func (check *checker) overloadTokenToName(op token.Token) string {
	switch op {
	case token.ADD:
		return "Add"
	case token.SUB:
		return "Subtract"
	case token.MUL:
		return "Multiply"
	case token.QUO:
		return "Divide"
	case token.REM:
		return "Modulo"
	case token.AND:
		return "BitAnd"
	case token.OR:
		return "BitOr"
	case token.SHL:
		return "BitShiftLeft"
	case token.SHR:
		return "BitShiftRight"
	case token.AND_NOT:
		return "BitAndNot"
	case token.XOR:
		return "BitXor"
	case token.LAND:
		return "And"
	case token.LOR:
		return "Or"
	case token.EQL:
		return "Equal"
	case token.LSS:
		return "Less"
	case token.GTR:
		return "Greater"
	case token.NOT:
		return "Not"
	case token.NEQ:
		return "NotEqual"
	case token.LEQ:
		return "LessOrEqual"
	case token.GEQ:
		return "GreaterOrEqual"
	default:
		return ""
	}
}

func (check *checker) overloadIsAddressable(recv *operand) bool {
	if recv.mode == variable {
		return true
	}

	switch recv.expr.(type) {
	case *ast.SliceExpr, *ast.CompositeLit, *ast.StarExpr:
		return true
	}

	return false
}

func (check *checker) overloadOperandType(oper *operand) Type {
	if oper == nil {
		return Typ[Invalid]
	}

	return oper.typ
}

func (check *checker) withIgnoredErrors(f func()) bool {
	err := check.firstErr
	errorf := check.conf.Error

	ret := true

	check.conf.Error = func(error) {
		ret = false
	}

	f()

	check.firstErr = err
	check.conf.Error = errorf

	return ret
}

func (check *checker) overloadLookupType(recv Type, oper *operand, name string) (*Func, Type) {
	ms := recv.MethodSet()

	for i := 0; i < ms.Len(); i++ {
		sel := ms.At(i)

		if sel.Kind() != MethodVal {
			continue
		}

		f := sel.Obj().(*Func)

		if !strings.HasPrefix(f.Name(), name) {
			continue
		}

		sig := f.Type().(*Signature)

		if sig.IsVariadic() {
			continue
		}

		if sig.Results().Len() != 1 {
			continue
		}

		params := sig.Params()

		if oper != nil {
			if params.Len() != 1 {
				continue
			}

			pt := params.At(0)

			if isUntyped(oper.typ) {
				mode := oper.mode

				ok := check.withIgnoredErrors(func() {
					check.convertUntyped(oper, pt.Type())
				})

				if !ok {
					oper.mode = mode
					continue
				}
			} else if pt.Type().String() != oper.typ.String() {
				continue
			}
		} else if params.Len() != 0 {
			continue
		}

		return f, sig.Results().At(0).Type()
	}

	return nil, Typ[Invalid]
}

func (check *checker) overloadLookupAddressable(recv *operand, oper *operand, name string) (*Func, Type) {
	if _, ok := recv.typ.(*Pointer); !ok {
		if check.overloadIsAddressable(recv) {
			return check.overloadLookupType(NewPointer(recv.typ), oper, name)
		}
	}

	return nil, Typ[Invalid]
}

func (check *checker) overloadLookup(recv *operand, oper *operand, name string) (*Func, Type) {
	if isUntyped(recv.typ) {
		return nil, Typ[Invalid]
	}

	// Try addressable first
	f, tp := check.overloadLookupAddressable(recv, oper, name)

	if f == nil {
		f, tp = check.overloadLookupType(recv.typ, oper, name)
	}

	return f, tp
}

func (check *checker) overloadBinaryOperator(x, y *operand, e ast.Expr) bool {
	xuntyp := isUntyped(x.typ)
	yuntyp := isUntyped(y.typ)

	if xuntyp && yuntyp {
		return false
	}

	be, ok := e.(*ast.BinaryExpr)

	if !ok {
		return false
	}

	n := check.overloadTokenToName(be.Op)

	if len(n) == 0 {
		return false
	}

	// Lookup method name on x satisfying type of y
	recv := x
	oper := y

	f, typ := check.overloadLookup(recv, oper, "Op_"+n)

	if f == nil {
		// Pre maybe?
		oper, recv = recv, oper

		f, typ = check.overloadLookup(recv, oper, "Op_Pre"+n)

		if f == nil {
			return false
		}
	}

	check.pkg.overloads[e] = OverloadInfo{
		Func: f,
		Recv: recv.expr,
		Oper: oper.expr,
	}

	// Resultant type would be of the result
	x.typ = typ
	x.mode = value

	return true
}

func (check *checker) overloadUnaryOperator(x *operand, e ast.Expr) bool {
	if isUntyped(x.typ) {
		return false
	}

	ue, ok := e.(*ast.UnaryExpr)

	if !ok {
		return false
	}

	n := check.overloadTokenToName(ue.Op)

	if len(n) == 0 {
		return false
	}

	// Lookup method name on x satisfying type of y
	f, typ := check.overloadLookup(x, nil, "Op_"+n)

	if f == nil {
		return false
	}

	check.pkg.overloads[e] = OverloadInfo{
		Func: f,
		Recv: x.expr,
		Oper: nil,
	}

	// Resultant type would be of the result
	x.typ = typ
	x.mode = value

	return true
}
