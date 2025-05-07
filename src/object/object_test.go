package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringHashKey(t *testing.T) {
	assert := assert.New(t)
	same1 := &String{Value: "Hello World"}
	same2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	assert.Equal(same1.HashKey(), same2.HashKey())
	assert.Equal(diff1.HashKey(), diff2.HashKey())
	assert.NotEqual(same1.HashKey(), diff2.HashKey())
}

func TestIntegerHashKey(t *testing.T) {
	assert := assert.New(t)
	same1 := &Integer{Value: 1}
	same2 := &Integer{Value: 1}
	diff1 := &Integer{Value: 2}
	diff2 := &Integer{Value: 2}

	assert.Equal(same1.HashKey(), same2.HashKey())
	assert.Equal(diff1.HashKey(), diff2.HashKey())
	assert.NotEqual(same1.HashKey(), diff2.HashKey())
}

func TestBooleanHashKey(t *testing.T) {
	assert := assert.New(t)
	same1 := &Boolean{Value: true}
	same2 := &Boolean{Value: true}
	diff1 := &Boolean{Value: false}
	diff2 := &Boolean{Value: false}

	assert.Equal(same1.HashKey(), same2.HashKey())
	assert.Equal(diff1.HashKey(), diff2.HashKey())
	assert.NotEqual(same1.HashKey(), diff2.HashKey())
}
