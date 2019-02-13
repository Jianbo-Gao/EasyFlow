// Author: Jianbo-Gao
// Define flags for taint analysis.

package vm

const SAFE_FLAG int = 0
const VALUE_FLAG int = 1 << 30

var global_taint_flag = SAFE_FLAG
