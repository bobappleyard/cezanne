package backend

// func TestTailCall(t *testing.T) {
// 	b := method{
// 		varc: 3,
// 		steps: []step{
// 			intStep{val: 1, into: 0},
// 			intStep{val: 1, into: 1},
// 			callStep{object: 0, method: "add", into: 2, params: []variable{1}},
// 			returnStep{val: 2},
// 		},
// 	}

// 	w := &assembler{}

// 	w.writeBlock(b)

// 	expect := new(assembly.Writer)
// 	expect.Natural(expect.Fixed(1))
// 	expect.Store(2)
// 	expect.Natural(expect.Fixed(1))
// 	expect.Store(3)
// 	expect.Load(2)
// 	expect.Store(5)
// 	expect.Load(3)
// 	expect.Store(6)
// 	expect.Load(6)
// 	expect.Store(2)
// 	expect.Load(5)
// 	expect.Call(expect.Method("add"), 0)

// 	assert.Equal(t, &w.dest, expect)
// }

// func TestPackage(t *testing.T) {

// }
