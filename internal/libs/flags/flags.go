package flags

import (
	"flag"
	"os"
	"time"
)

// Value 表示flag值的类型约束.
type Value interface {
	int | int64 | uint | uint64 | float64 | time.Duration | bool | string
}

var (
	// flagSet 表示flag集.
	flagSet *flag.FlagSet

	// valueMap 存储flag值的map.
	valueMap = map[string]any{}
)

// getFlagSet 获取flag集.
func getFlagSet() *flag.FlagSet {
	if flagSet == nil {
		if len(os.Args) == 0 {
			flagSet = flag.NewFlagSet("", flag.ExitOnError)
		} else {
			flagSet = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		}
	}
	return flagSet
}

// addValue 添加flag值.
func addValue[val Value](name string, p *val) {
	if valueMap == nil {
		valueMap = make(map[string]any)
	}
	valueMap[name] = p
}

// Parse 解析flag.
func Parse() {
	if flagSet == nil {
		return
	}
	flagSet.Parse(os.Args[1:])
}

// Clear 清除所有flag.
func Clear() {
	flagSet = nil
	valueMap = nil
}

// Int 设置int类型flag.
func Int(name string, value int, usage string) *int {
	pv := getFlagSet().Int(name, value, usage)
	addValue(name, pv)
	return pv
}

// Int64 设置int64类型flag.
func Int64(name string, value int64, usage string) *int64 {
	pv := getFlagSet().Int64(name, value, usage)
	addValue(name, pv)
	return pv
}

// Uint 设置uint类型flag.
func Uint(name string, value uint, usage string) *uint {
	pv := getFlagSet().Uint(name, value, usage)
	addValue(name, pv)
	return pv
}

// Uint64 设置uint64类型flag.
func Uint64(name string, value uint64, usage string) *uint64 {
	pv := getFlagSet().Uint64(name, value, usage)
	addValue(name, pv)
	return pv
}

// Float64 设置float64类型flag.
func Float64(name string, value float64, usage string) *float64 {
	pv := getFlagSet().Float64(name, value, usage)
	addValue(name, pv)
	return pv
}

// Duration 设置time.Duration类型flag.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	pv := getFlagSet().Duration(name, value, usage)
	addValue(name, pv)
	return pv
}

// Bool 设置bool类型flag.
func Bool(name string, value bool, usage string) *bool {
	pv := getFlagSet().Bool(name, value, usage)
	addValue(name, pv)
	return pv
}

// String 设置string类型flag.
func String(name string, value string, usage string) *string {
	pv := getFlagSet().String(name, value, usage)
	addValue(name, pv)
	return pv
}

// GetValue 获取flag值.
func GetValue[val Value](name string) (v val, exist bool) {
	if pv := valueMap[name]; pv == nil {
		exist = false
	} else {
		v = *(pv.(*val))
		exist = true
	}
	return
}
