package parser

import (
	"fmt"
	"testing"

	"monkey/ast"
	"monkey/lexer"

	"github.com/stretchr/testify/assert"
)

func TestLetStatements(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {

		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)
		assert.NotNil(program)

		assert.Equal(1, len(program.Statements))

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(assert, val, tt.expectedValue) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got=%s",
			name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return 10", 10},
		{"return 993322;", 993322},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(1, len(program.Statements))

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		assert.True(ok)

		assert.Equal("return", returnStmt.TokenLiteral())

		testLiteralExpression(assert, returnStmt.ReturnValue, tt.expectedValue)
	}
}

func TestIfExpression(t *testing.T) {
	assert := assert.New(t)
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(1, len(program.Statements))

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])

	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(ok, "stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)

	testInfixExpression(assert, exp.Condition, "x", "<", "y")

	assert.Equal(1, len(exp.Then.Statements))
	then, ok := exp.Then.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok)
	testIdentifier(assert, then.Expression, "x")

	assert.Nil(exp.Else)
}

func TestIfElseExpression(t *testing.T) {
	assert := assert.New(t)
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(1, len(program.Statements))

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])

	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(ok, "stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)

	testInfixExpression(assert, exp.Condition, "x", "<", "y")

	assert.Equal(1, len(exp.Then.Statements))
	then, ok := exp.Then.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok)
	testIdentifier(assert, then.Expression, "x")

	assert.NotNil(exp.Else)
	alt, ok := exp.Else.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok)
	testIdentifier(assert, alt.Expression, "y")
}

func TestFunctionLiteralParsing(t *testing.T) {
	assert := assert.New(t)
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(len(program.Statements), 1, "program has not enough statements")

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement")

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	assert.True(ok, "stmt.Expression is not ast.FunctionLiteral. got=%T", stmt.Expression)

	assert.Equal(2, len(function.Parameters), "wrong function literal parameter count")

	testLiteralExpression(assert, function.Parameters[0], "x")
	testLiteralExpression(assert, function.Parameters[1], "y")

	assert.Equal(1, len(function.Body.Statements), "function body has not 1 statement")

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok, "function body statement is not ast.ExpressionStatement")

	testInfixExpression(assert, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		assert.Equal(len(tt.expectedParams), len(function.Parameters))

		for i, ident := range tt.expectedParams {
			testLiteralExpression(assert, function.Parameters[i], ident)
		}
	}
}

func TestCallExpression(t *testing.T) {
	assert := assert.New(t)
	input := "add(1, 2 * 3, 4 + 5);"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(1, len(program.Statements))

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok)

	exp, ok := stmt.Expression.(*ast.CallExpression)
	assert.True(ok)

	if !testIdentifier(assert, exp.Function, "add") {
		return
	}

	assert.Equal(3, len(exp.Arguments))

	testLiteralExpression(assert, exp.Arguments[0], 1)
	testInfixExpression(assert, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(assert, exp.Arguments[2], 4, "+", 5)
}

func TestIdentifierExpression(t *testing.T) {
	assert := assert.New(t)
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(1, len(program.Statements), "program has not enough statements")
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement")

	testIdentifier(assert, stmt.Expression, "foobar")
}

func TestIntegerLiteralExpression(t *testing.T) {
	assert := assert.New(t)
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Equal(1, len(program.Statements), "program has not enough statements")
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement")

	testLiteralExpression(assert, stmt.Expression, int64(5))
}

func TestBooleanExpression(t *testing.T) {
	assert := assert.New(t)

	input := []struct {
		input   string
		literal bool
	}{
		{"false;", false},
		{"true;", true},
	}

	for _, tt := range input {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(1, len(program.Statements), "program has not enough statements")
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement")

		testLiteralExpression(assert, stmt.Expression, tt.literal)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	assert := assert.New(t)
	input := `"hello world"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	assert.True(ok, "exp not *ast.StringLiteral, got=%T", stmt.Expression)
	assert.Equal("hello world", literal.Value)
}

func TestParsingPrefixExpression(t *testing.T) {
	assert := assert.New(t)
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(1, len(program.Statements), "program has not enough statements")

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement")

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(ok, "exp not *ast.PrefixExpression")
		assert.Equal(tt.operator, exp.Operator)

		testLiteralExpression(assert, exp.Right, tt.value)
	}
}

func TestParsingInfixExpression(t *testing.T) {
	assert := assert.New(t)
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(1, len(program.Statements), "program has not enough statements")
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(ok, "program.Statements[0] is not ast.ExpressionStatement")

		testInfixExpression(assert, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue)
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	assert := assert.New(t)
	input := "[1, 2 * 2, 3 + 3]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	assert.True(ok, "exp not ast.ArrayLiteral. got=%T", stmt.Expression)
	assert.Equal(3, len(array.Elements))

	testIntegerLiteral(assert, array.Elements[0], 1)
	testInfixExpression(assert, array.Elements[1], 2, "*", 2)
	testInfixExpression(assert, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	assert := assert.New(t)
	input := "myArray[1 + 1]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	assert.True(ok, "exp not *ast.IndexExpression. got=%T", stmt.Expression)
	testIdentifier(assert, indexExp.Left, "myArray")
	testInfixExpression(assert, indexExp.Index, 1, "+", 1)
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		assert.Equal(tt.expected, actual)
	}
}

func testLiteralExpression(assert *assert.Assertions, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(assert, exp, int64(v))
	case int64:
		return testIntegerLiteral(assert, exp, v)
	case bool:
		return testBooleanLiteral(assert, exp, v)
	case string:
		return testIdentifier(assert, exp, v)
	}
	return false
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	assert := assert.New(t)
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	assert.True(ok, "exp is not ast.HashLiteral. got=%T", stmt.Expression)

	assert.Equal(3, len(hash.Pairs))

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		assert.True(ok)
		testIntegerLiteral(assert, value, expected[literal.String()])
	}
}

func TestParsingHashLiteralsIntegerKeys(t *testing.T) {
	assert := assert.New(t)
	input := `{1: "one", 2: "two", 3: "three"}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	assert.True(ok, "exp is not ast.HashLiteral. got=%T", stmt.Expression)

	assert.Equal(3, len(hash.Pairs))

	expected := map[int64]string{
		1: "one",
		2: "two",
		3: "three",
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.IntegerLiteral)
		assert.True(ok)
		testStringLiteral(assert, value, expected[literal.Value])
	}
}

func TestParsingHashLiteralsBooleanKeys(t *testing.T) {
	assert := assert.New(t)
	input := `{true: 1, false: 2}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	assert.True(ok, "exp is not ast.HashLiteral. got=%T", stmt.Expression)

	assert.Equal(2, len(hash.Pairs))

	expected := map[bool]int64{
		true:  1,
		false: 2,
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.Boolean)
		assert.True(ok)
		testIntegerLiteral(assert, value, expected[literal.Value])
	}
}

func testParseEmptyHashLiteral(t *testing.T) {
	assert := assert.New(t)
	input := `{}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	assert.True(ok, "exp is not ast.HashLiteral. got=%T", stmt.Expression)
	assert.Equal(0, len(hash.Pairs))
}

func testParsingHashLiteralsWithExpressions(t *testing.T) {
	assert := assert.New(t)
	input := `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	assert.True(ok, "stmt is not ast.HashLiteral. got=%T", stmt.Expression)
	assert.Equal(3, len(hash.Pairs))

	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(assert, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(assert, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(assert, e, 15, "/", 5)
		},
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		assert.True(ok)
		testFunc, ok := tests[literal.String()]
		assert.True(ok)
		testFunc(value)
	}
}

func testInfixExpression(assert *assert.Assertions, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.InfixExpression)
	assert.True(ok, "exp is not an ast.InfixExpression. got=%T(%s)", exp, exp)

	if !testLiteralExpression(assert, opExp.Left, left) {
		return false
	}

	assert.Equal(operator, opExp.Operator)

	if !testLiteralExpression(assert, opExp.Right, right) {
		return false
	}

	return true
}

func testIdentifier(assert *assert.Assertions, exp ast.Expression, value string) bool {
	identifier, ok := exp.(*ast.Identifier)
	assert.True(ok, "exp not *ast.Identifier. got=%T", exp)
	assert.Equal(value, identifier.Value)
	assert.Equal(value, identifier.TokenLiteral())
	return true
}

func testIntegerLiteral(assert *assert.Assertions, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	assert.True(ok, "exp not *ast.IntegerLiteral. got=%T", il)
	assert.Equal(value, integ.Value)
	assert.Equal(fmt.Sprintf("%d", value), integ.TokenLiteral())
	return true
}

func testBooleanLiteral(assert *assert.Assertions, b ast.Expression, value bool) bool {
	boolean, ok := b.(*ast.Boolean)
	assert.True(ok, "exp not *ast.Boolean. got=%T", b)
	assert.Equal(value, boolean.Value)
	assert.Equal(fmt.Sprintf("%t", value), boolean.TokenLiteral())
	return true
}

func testStringLiteral(assert *assert.Assertions, l ast.Expression, value string) bool {
	lit, ok := l.(*ast.StringLiteral)
	assert.True(ok, "exp not *ast.StringLiteral. got=%T", l)
	assert.Equal(value, lit.Value)
	assert.Equal(fmt.Sprintf("%s", value), lit.TokenLiteral())
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser had %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
