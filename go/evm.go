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
var power256 = new(big.Int).Lsh(big.NewInt(1), 256)                                    // max value of uint256
// Run runs the EVM code and returns the stack and a success indicator.
func Evm(code []byte) ([]*big.Int, bool) {
	var a, b, c *big.Int // For top 2 values always

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
				z = b
			} else {
				z = wrap(new(big.Int).Div(a, b))
			}
			stack = push(stack, z)
		case 0x06: //MOD
			stack, a, b = pop2(stack)
			z := MOD(a, b)
			stack = push(stack, z)

		case 0x08: //ADDMOD
			stack, a, b, c = pop3(stack)
			sum := wrap(new(big.Int).Add(a, b))
			z := MOD(sum, c)
			stack = push(stack, z)
		case 0x09: //MULMOD
			stack, a, b, c = pop3(stack)
			product := (new(big.Int).Mul(a, b))
			z := MOD(product, c)
			stack = push(stack, z)
		case 0x0a: //EXP
			stack, a, b = pop2(stack)
			z := new(big.Int).Exp(a, b, power256)
			stack = push(stack, z)
		case 0x0b: // SIGNEXTEND
			stack, a, b = pop2(stack)

			// If a >= 32 then there's no change
			if a.Cmp(big.NewInt(31)) >= 0 {
				stack = push(stack, b)
				continue
			}

			// The bit position to sign extend from is (a+1)*8-1
			signBitPos := new(big.Int).Mul(new(big.Int).Add(a, big.NewInt(1)), big.NewInt(8))
			signBitPos = new(big.Int).Sub(signBitPos, big.NewInt(1))

			// Check if the sign bit is set
			if b.Bit(int(signBitPos.Int64())) == 1 {
				// The sign bit is 1, so set all higher bits to 1
				mask := new(big.Int).Lsh(big.NewInt(1), uint(signBitPos.Uint64()+1))
				mask = new(big.Int).Sub(mask, big.NewInt(1))
				mask = new(big.Int).Not(mask)

				// Apply mask to set all higher bits to 1
				b = wrap(new(big.Int).Or(b, mask))
			} else {
				// The sign bit is 0, so clear all higher bits
				mask := new(big.Int).Lsh(big.NewInt(1), uint(signBitPos.Uint64()+1))
				mask = new(big.Int).Sub(mask, big.NewInt(1))

				// Apply mask to clear all higher bits
				b = wrap(new(big.Int).And(b, mask))
			}

			stack = push(stack, b)
		case 0x05: // SDIV
			stack, a, b = pop2(stack)
			z := SDIV(a, b)
			stack = push(stack, wrap(z))
		case 0x07: // SMOD
			stack, a, b = pop2(stack)
			var z *big.Int
			if b.Cmp(big.NewInt(0)) == 0 {
				z = b
			} else {
				z = SDIV(a, b)                                             // a//b
				z = new(big.Int).Sub(normalise(a), new(big.Int).Mul(z, b)) // a - b(a//b)
			}
			stack = push(stack, wrap(z))
		case 0x10: // LT (a<b ==> 1 )
			stack, a, b = pop2(stack)
			if a.Cmp(b) == -1 {
				stack = push(stack, big.NewInt(1))
			} else {
				stack = push(stack, big.NewInt(0))
			}
		case 0x11: // GT (a>b ==> 1 )
			stack, a, b = pop2(stack)
			if a.Cmp(b) == 1 {
				stack = push(stack, big.NewInt(1))
			} else {
				stack = push(stack, big.NewInt(0))
			}
		case 0x12: // SLT (a<b ==> 1 )
			stack, a, b = pop2(stack)
			if normalise(a).Cmp(normalise(b)) == -1 {
				stack = push(stack, big.NewInt(1))
			} else {
				stack = push(stack, big.NewInt(0))
			}
		case 0x13: // GT (a>b ==> 1 )
			stack, a, b = pop2(stack)
			if normalise(a).Cmp(normalise(b)) == 1 {
				stack = push(stack, big.NewInt(1))
			} else {
				stack = push(stack, big.NewInt(0))
			}
		case 0x14: // EQ (a=b ==> 1 )
			stack, a, b = pop2(stack)
			if normalise(a).Cmp(normalise(b)) == 0 {
				stack = push(stack, big.NewInt(1))
			} else {
				stack = push(stack, big.NewInt(0))
			}
		case 0x15: // ISZERO 
			stack, a= pop(stack)
			if a.Cmp(big.NewInt(0)) == 0 {
				stack = push(stack, big.NewInt(1))
			} else {
				stack = push(stack, big.NewInt(0))
			}
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

func pop3(stack []*big.Int) ([]*big.Int, *big.Int, *big.Int, *big.Int) {
	var a, b, c *big.Int
	stack, a = pop(stack)
	stack, b = pop(stack)
	stack, c = pop(stack)
	return stack, a, b, c
}

func MOD(a *big.Int, b *big.Int) *big.Int {
	var z *big.Int
	if b.Cmp(big.NewInt(0)) == 0 {
		z = b
	} else {
		z = wrap(new(big.Int).Mod(a, b))
	}
	return z
}

func bitcut(b *big.Int, bitpos *big.Int) *big.Int {
	b = new(big.Int).Lsh(b, uint(256-bitpos.Uint64()))
	b = new(big.Int).Rsh(b, uint(256-bitpos.Uint64()))
	return b
}

func normalise(a *big.Int) *big.Int {
	if a.Bit(255) == 1 {
		return new(big.Int).Mul(new(big.Int).Sub(power256, a), big.NewInt(-1))
	}
	return a
}

func SDIV(a *big.Int, b *big.Int) *big.Int {
	z := new(big.Int)
	if b.Cmp(big.NewInt(0)) == 0 {
		z = b
	} else {
		z = new(big.Int).Div(normalise(a), normalise(b))
	}
	return z
}
