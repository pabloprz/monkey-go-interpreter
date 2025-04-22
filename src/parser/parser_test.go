package parser

import (
	"fmt"
	"testing"

	"monkey/ast"
	"monkey/lexer"

	"github.com/stretchr/testify/assert"
)

func TestLetStatements(t *testing.T) {
	input := `
	let x = 5;
	let y = 10;
	let foobar = 838383;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
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
	input := `
	return 5;
	return 10;
	return 993322;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got %q", returnStmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return'. got %q", returnStmt.TokenLiteral())
		}
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
