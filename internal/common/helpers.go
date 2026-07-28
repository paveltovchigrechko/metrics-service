package common

func Ptr[T any](v T) *T {
	return &v
}

func Int64Ptr(v int64) *int64       { return &v }
func Float64Ptr(v float64) *float64 { return &v }
