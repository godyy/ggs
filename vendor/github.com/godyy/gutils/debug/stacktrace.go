package debug

import (
	"runtime"
	"strconv"
	"strings"
)

// StackTrace 更轻量的堆栈捕获：使用 Callers + FuncForPC，避免 CallersFrames。
// skip: 跳过的层级（内部再额外跳过本函数和 runtime.Callers 两层，总体等价 skip+2）。
// max: 最大帧数（<=0 则默认 32）。
func StackTrace(skip, max int) string {
	if max <= 0 {
		max = 32
	}
	pcs := make([]uintptr, max)
	n := runtime.Callers(skip+2, pcs)
	if n == 0 {
		return ""
	}

	var b strings.Builder
	// 逐帧解析，避免构造中间切片
	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pc)

		// 写入：func\n\tfile:line\n
		b.WriteString(fn.Name())
		b.WriteString("\n\t")
		b.WriteString(file)
		b.WriteString(":")
		// 避免 fmt，减少临时分配
		b.WriteString(strconv.FormatInt(int64(line), 10))
		b.WriteString("\n")
	}
	return b.String()
}
