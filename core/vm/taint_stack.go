// Author: Jianbo-Gao
// Recording EVM stack for taint analysis.

package vm

import "fmt"
import "encoding/json"

type TaintStack struct {
	data []int
}

func newtaintstack() *TaintStack {
	return &TaintStack{data: make([]int, 0, 1024)}
}

func (st *TaintStack) Data() []int {
	return st.data
}

func (st *TaintStack) push(d int) {
	// NOTE push limit (1024) is checked in baseCheck
	//stackItem := new(big.Int).Set(d)
	//st.data = append(st.data, stackItem)
	st.data = append(st.data, d)
}
func (st *TaintStack) pushN(ds ...int) {
	st.data = append(st.data, ds...)
}

func (st *TaintStack) pop() (ret int) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *TaintStack) len() int {
	return len(st.data)
}

func (st *TaintStack) swap(n int) {
	st.data[st.len()-n], st.data[st.len()-1] = st.data[st.len()-1], st.data[st.len()-n]
}

// func (st *TaintStack) dup(pool *intPool, n int) {
// 	st.push(pool.get().Set(st.data[st.len()-n]))
// }

func (st *TaintStack) dup(n int) {
	// TODO: intPool
	st.push(st.data[st.len()-n])
}

func (st *TaintStack) peek() int {
	return st.data[st.len()-1]
}

// Back returns the n'th item in stack
func (st *TaintStack) Back(n int) int {
	return st.data[st.len()-n-1]
}

func (st *TaintStack) require(n int) error {
	if st.len() < n {
		return fmt.Errorf("stack underflow (%d <=> %d)", len(st.data), n)
	}
	return nil
}

func (st *TaintStack) Print() {
	fmt.Println("### stack ###")
	if len(st.data) > 0 {
		for i, val := range st.data {
			fmt.Printf("%-3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
}

func (st *TaintStack) JPrint() {
	if j_data, err := json.Marshal(st.data); err == nil {
		fmt.Printf("TaintStack:%s\n", j_data)
	}
}

func main() {
	var t_stack TaintStack
	t_stack.data = append(t_stack.data, 1)
	t_stack.data = append(t_stack.data, 2)
	t_stack.data = append(t_stack.data, 3)
	t_stack.Print()
	print("pop: ")
	println(t_stack.pop())
	t_stack.Print()
}
