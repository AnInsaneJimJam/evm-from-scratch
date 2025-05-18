// Package evm is an **incomplete** implementation of the Ethereum Virtual
// Machine for the "EVM From Scratch" course:
// https://github.com/w1nt3r-eth/evm-from-scratch
//
// To work on EVM From Scratch In Go:
//
// - Install Golang: https://golang.org/doc/install
// - Go to the `go` directory: `cd go`
// - Edit `evm.go` (this file!), see TODO below
// - Run `go test ./...` to run the tests
package evm

import (
	"math/big"
)

var maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)) // max value of uint256
// Run runs the EVM code and returns the stack and a success indicator.
func Evm(code []byte) ([]*big.Int, bool) {
	var a, b *big.Int // For top 2 values always

	var stack []*big.Int
	pc := 0

	for pc < len(code) {
		op := code[pc]
		pc++

		if op >= 0x60 && op <= 0x7f { // All PUSH => PUSH1 .... PUSH32
			num := int(op - 0x60)
			if pc >= (len(code) - num) {
				return reverse(stack), false
			}
			value := new(big.Int).SetBytes(code[pc : pc+num+1]) // Big-endian
			stack = push(stack, value)
			pc += num + 1
			continue
		}
		switch op {
		case 0x00: // STOP
			return stack, true
		case 0x5f: // PUSH0
			stack = push(stack, big.NewInt(0))
		case 0x50: // POP
			stack, _ = pop(stack)
		case 0x01: //ADD
			stack, a, b = pop2(stack)
			z := wrap(new(big.Int).Add(a, b))
			stack = push(stack, z)
		case 0x02: //MUL
			stack, a, b = pop2(stack)
			z := wrap(new(big.Int).Mul(a, b))
			stack = push(stack, z)
		case 0x03: //SUb
			stack, a, b = pop2(stack)
			z := wrap(new(big.Int).Sub(a, b))
			stack = push(stack, z)
		case 0x04: //DIV
			stack, a, b = pop2(stack)
			var z *big.Int
			if b.Cmp(big.NewInt(0)) == 0 {
				z = b;
			}else{
			z = wrap(new(big.Int).Div(a, b))}
			stack = push(stack, z)
		}
	}
	return reverse(stack), true
}

func reverse(stack []*big.Int) []*big.Int {
	n := len(stack)
	out := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		out[i] = stack[n-1-i]
	}
	return out
}

func pop(stack []*big.Int) ([]*big.Int, *big.Int) {
	n := len(stack)
	elem := stack[n-1]
	out := stack[0 : n-1]
	return out, elem
}

func push(stack []*big.Int, elem *big.Int) []*big.Int {
	stack = append(stack, elem)
	return stack
}

func pop2(stack []*big.Int) ([]*big.Int, *big.Int, *big.Int) {
	var a, b *big.Int
	stack, a = pop(stack)
	stack, b = pop(stack)
	return stack, a, b
}

func wrap(z *big.Int) *big.Int {
	bits := z.BitLen()
	if bits > 256 {
		z = new(big.Int).And(z, maxUint256)
	}
	if z.Cmp(big.NewInt(0)) == -1 {
		a := new(big.Int).Add(maxUint256, big.NewInt(1))
		z = new(big.Int).Add(a, z)
	}
	return z
}
