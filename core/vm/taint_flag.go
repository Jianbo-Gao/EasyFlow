// Author: Jianbo-Gao
// Define flags for taint analysis.

package vm

const SAFE_FLAG int = 0
const CALLDATA_FLAG int = 1
const POTENTIAL_OVERFLOW_FLAG int = 1 << 1
const PROTECTED_OVERFLOW_FLAG int = 1 << 2
const OVERFLOW_FLAG int = 1 << 3

var global_taint_flag = SAFE_FLAG
