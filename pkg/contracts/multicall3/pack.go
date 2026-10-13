package multicall3

import (
	"encoding/binary"
)

const abiWord = 32

// packCalls ABI-encodes the arguments of tryAggregate and tryBlockAndAggregate,
// (bool requireSuccess, Call[] calls), after selector.
//
// Every call carries an offset to its call data rather than the data itself,
// so calls with the same call data point at one copy of it. Reading one
// account's balance of many tokens repeats the call data in every call: one
// copy halves the request. The contract reads the encoding like any other.
func packCalls(selector []byte, requireSuccess bool, calls []IMulticall3Call) []byte {
	n := len(calls)
	// Each distinct call data, in the order it first appears, and where it
	// sits from the start of the array body.
	datas := make([][]byte, 0, 1)
	position := make(map[string]int, 1)
	next := n*abiWord + n*2*abiWord
	for _, call := range calls {
		if _, seen := position[string(call.CallData)]; seen {
			continue
		}
		position[string(call.CallData)] = next
		datas = append(datas, call.CallData)
		next += abiWord + padded(len(call.CallData))
	}

	out := make([]byte, 0, len(selector)+3*abiWord+next)
	out = append(out, selector...)
	out = appendBool(out, requireSuccess)
	out = appendUint(out, 2*abiWord) // where the array starts
	out = appendUint(out, n)
	for i := range calls {
		out = appendUint(out, tupleAt(n, i))
	}
	for i, call := range calls {
		out = append(out, make([]byte, abiWord-len(call.Target))...)
		out = append(out, call.Target[:]...)
		out = appendUint(out, position[string(call.CallData)]-tupleAt(n, i))
	}
	for _, data := range datas {
		out = appendUint(out, len(data))
		out = append(out, data...)
		out = append(out, make([]byte, padded(len(data))-len(data))...)
	}
	return out
}

// tupleAt is where call i of n sits from the start of the array body: past
// the n offsets, each call taking a word for its target and one for the
// offset to its call data.
func tupleAt(n, i int) int {
	return n*abiWord + i*2*abiWord
}

func padded(length int) int {
	return (length + abiWord - 1) / abiWord * abiWord
}

func appendUint(out []byte, v int) []byte {
	var word [abiWord]byte
	binary.BigEndian.PutUint64(word[abiWord-8:], uint64(v))
	return append(out, word[:]...)
}

func appendBool(out []byte, v bool) []byte {
	if v {
		return appendUint(out, 1)
	}
	return appendUint(out, 0)
}
