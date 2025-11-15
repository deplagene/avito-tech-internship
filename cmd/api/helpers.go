package api

// Ptr возвращает указатель на переданное значение
func Ptr[T any](v T) *T {
	return &v
}
