package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalIntegerExpression(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(assert, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{`"test" == "test"`, true},
		{`"test" == " test"`, false},
		{`"12345" == "12345"`, true},
		{`"test" != "test"`, false},
		{`"test" != " test"`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(assert, evaluated, tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	assert := assert.New(t)
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	assert.True(ok, "object is not String. got=%T (%+v)", evaluated, evaluated)
	assert.Equal("Hello World!", str.Value)
}

func TestStringConcatenation(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input    string
		expected string
	}{
		{`"Hello" + " " + "World!"`, "Hello World!"},
		{`"1" * 3`, "111"},
		{`"abc" * 0"`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		assert.True(ok, "object is not String. got=%T (%+v)", evaluated, evaluated)
		assert.Equal(tt.expected, str.Value)
	}
}

func TestBangOperator(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(assert, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(assert, evaluated, int64(integer))
		} else {
			testNullObject(assert, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9", 10},
		{"return 2 * 5; 9", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }

  return 1;
}
`,
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(assert, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(assert, testEval(tt.input), tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return true + false;
  }

  return 1;
}
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"1" * -3`,
			"negative argument error: STRING * -3",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tt := range tests {
		assert := assert.New(t)
		evaluated := testEval(tt.input)

		errorObj, ok := evaluated.(*object.Error)
		assert.True(ok, "evaluated was not an error, got %T(%+v)", evaluated, evaluated)
		assert.Equal(tt.expectedMessage, errorObj.Message)
	}
}

func TestFunctionObject(t *testing.T) {
	assert := assert.New(t)
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	assert.True(ok, "evaluated was not a function, got %T(%+v)", evaluated, evaluated)

	assert.Equal(1, len(fn.Parameters))
	assert.Equal("x", fn.Parameters[0].String())
	assert.Equal("(x + 2)", fn.Body.String())
}

func TestFunctionApplication(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(assert, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	assert := assert.New(t)
	input := `
let newAdder = fn(x) {
  fn(y) { x + y };
};

let addTwo = newAdder(2);
addTwo(2);
	`
	testIntegerObject(assert, testEval(input), 4)
}

func TestArrayLiterals(t *testing.T) {
	assert := assert.New(t)
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	assert.True(ok, "object is not Array. got=%T (%+v)", evaluated, evaluated)
	assert.Equal(3, len(result.Elements))

	testIntegerObject(assert, result.Elements[0], 1)
	testIntegerObject(assert, result.Elements[1], 4)
	testIntegerObject(assert, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(assert, evaluated, int64(integer))
		} else {
			testNullObject(assert, evaluated)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	assert := assert.New(t)
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	assert.True(ok, "Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	assert.Equal(len(expected), len(result.Pairs))

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		assert.True(ok, "no pair for given key in Pairs")

		testIntegerObject(assert, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(assert, evaluated, int64(integer))
		} else {
			testNullObject(assert, evaluated)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(assert, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			assert.True(ok, "object is not Error. got=%T (+%v)")
			assert.Equal(tt.expected, errObj.Message)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(assert *assert.Assertions, obj object.Object, expected int64) {
	result, ok := obj.(*object.Integer)
	assert.True(ok, "object is not Integer, got=%T (%+v)", obj, obj)
	assert.Equal(expected, result.Value, "object has wrong value")
}

func testBooleanObject(assert *assert.Assertions, obj object.Object, expected bool) {
	result, ok := obj.(*object.Boolean)
	assert.True(ok, "object is not Boolean, got=%T (%+v)", obj, obj)
	assert.Equal(expected, result.Value, "object has wrong value")
}

func testNullObject(assert *assert.Assertions, obj object.Object) {
	assert.Equal(obj, NULL)
}
