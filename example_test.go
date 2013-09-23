package main

import (
	"fmt"
)

type vec4 [4]float32

func (v vec4) Op_Multiply(o vec4) vec4 {
	return vec4{v[0] * o[0], v[1] * o[1], v[2] * o[2], v[3] * o[3]}
}

func (v vec4) Op_MultiplyScalar(o float32) vec4 {
	return vec4{v[0] * o, v[1] * o, v[2] * o, v[3] * o}
}

func (v vec4) Op_PreMultiplyScalar(o float32) vec4 {
	return v.Op_MultiplyScalar(o)
}

func (v vec4) Op_Add(o vec4) vec4 {
	return vec4{v[0] + o[0], v[1] + o[1], v[2] + o[2], v[3] + o[3]}
}

func (v vec4) Op_AddScalar(o float32) vec4 {
	return vec4{v[0] + o, v[1] + o, v[2] + o, v[3] + o}
}

func (v vec4) Op_PreAddScalar(o float32) vec4 {
	return v.Op_AddScalar(o)
}

func (v vec4) Op_Subtract(o vec4) vec4 {
	return vec4{v[0] - o[0], v[1] - o[1], v[2] - o[2], v[3] - o[3]}
}

func (v vec4) Op_SubtractScalar(o float32) vec4 {
	return vec4{v[0] - o, v[1] - o, v[2] - o, v[3] - o}
}

func (v vec4) Op_PreSubtractScalar(o float32) vec4 {
	return vec4{o - v[0], o - v[1], o - v[2], o - v[3]}
}

func ExampleOverload() {
	v1 := vec4{1, 2, 3, 4}
	v2 := vec4{5, 6, 7, 8}

	// Generates: ret := v1.Op_PreMultiplyScalar(2).Op_Multiply(v2).Op_Add(v1).Op_SubtractScalar(4)
	ret := 2*v1*v2 + v1 - 4

	fmt.Println(ret)
}
