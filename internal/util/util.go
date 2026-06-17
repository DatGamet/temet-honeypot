package util

func Pointer[T any](v T) *T {
	return &v
}

func InlineIfElse[V any](condition bool, ifTrue V, ifFalse V) V {
	if condition {
		return ifTrue
	}

	return ifFalse
}
