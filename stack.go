package glox

type stack[T any] struct {
	Push    func(T)
	Pop     func() T
	Length  func() int
	IsEmpty func() bool
	Peek	func() T
	Get		func(index int) T
}

func Stack[T any]() stack[T] {
	elements := make([]T, 0)
	return stack[T]{
		Push: func(ele T) {
			elements = append(elements, ele)
		},
		Pop: func() T {
			ele := elements[len(elements)-1]
			elements = elements[:len(elements)-1]
			return ele
		},
		Length: func() int {
			return len(elements)
		},
		IsEmpty: func() bool {
			return len(elements) == 0
		},
		Peek: func() T {
			return elements[len(elements)-1]
		},
		Get: func(index int) T {
			return elements[index]
		},
	}
}
