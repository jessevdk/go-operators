Disclaimer: *This is a hack!*

This project is an attempt to add operator overloading to the Go language. There
are very good reasons why Go doesn't support operator overloading as part of
the language and I fully support and agree with those reasons.

That being said, there are a few valid cases where it makes sense to have
operators defined for types other than the builtin types. One such valid
case, in my opinion, is for numeric types such as vectors or matrices. Of course,
Go was never designed for writing numeric heavy programs, so you could just
argue that Go is simply not the right tool for the job. Although I wouldn't
disagree with this, I still find Go an exceptionally enjoyable language and
recently I really wanted to write some vector-ish code.

go-operators allows you to add support for all Go operators to any custom
defined type, by adding appropriate methods on that type. The way this works is
that go-operators is a kind of preprocessor for your Go code which will parse
and rewrite your original source to map operators, on types which do not natively
support operators, to method calls on the operand(s) of the operator.

First, the original source is converted to an AST by using the builtin go/ast package.
Then, a patched version of go/types (available in go.tools) steps through the
AST and resolves the types of all the expressions in the AST. When an operator
is encountered which operators on non-numeric types, a method lookup is
performed on the first operand type. If the appropriate operator overloading
method can be found (with the correct operands type and return type), then the
AST node representing the operator is replaced with a method call. For binary
operators, if the first operand does not have an appropriate overloaded method,
the second operator is tested for a Pre- overload method. This way you can
overload both `v * 5.0` and `5.0 * v` (for example) on the type of `v`.

# Using go-operators
To use go-operators you will have to define special methods on your custom type.
These methods are prefixed with `Op_Multiply`, `Op_Add`, `Op_Subtract` etc.,
depending on the operator to overload. Any method which has such a prefix is
a potential candidate. Thus, if your type supports operators on various operand
types, then you can add methods such as `Op_MultiplyVec3`, `Op_MultiplyMat3`,
etc. with the appropriate argument types. The return type of the method can be
anything, but there must be extactly one return value. For overloaded binary
operators there must be exactly one argument while for overloaded unary operators
there must be exactly zero arguments.

After you have defined these methods, you can simply use
`go-operators --output OUTPUT_DIR SOURCE_DIR` to parse all go files in
`SOURCE_DIR` and generate code in `OUTPUT_DIR`. This should replace all operators,
where needed, with appropriate method calls.

# Not using go-operators
There is much to say for not using go-operators. I already mentioned that it's
a hack, right? Using go-operators relies on a preprocessing step which, although
fairly robust, is subject to bugs. The go/types project is still in development
and certain constructs are bound not to be correctly parsed (yet). Additionally,
using go-operators makes your project no longer go-gettable since a preprocessing
step needs to be performed before the code is actually usable.
