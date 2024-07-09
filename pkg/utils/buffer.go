package utils

// CircularBuffer represents a fixed-size queue that overwrites oldest elements when full
type CircularBuffer[T any] struct {
	buffer []T
	head   int
	tail   int
	size   int
	count  int
}

// NewCircularBuffer creates a new CircularBuffer with the given size
func NewCircularBuffer[T any](size int) *CircularBuffer[T] {
	return &CircularBuffer[T]{
		buffer: make([]T, size),
		size:   size,
	}
}

// Push adds an element to the front of the buffer
func (cb *CircularBuffer[T]) Push(value T) {
	cb.buffer[cb.tail] = value
	if cb.count == cb.size {
		cb.head = (cb.head + 1) % cb.size // Move head if buffer is full
	} else {
		cb.count++
	}
	cb.tail = (cb.tail + 1) % cb.size
}

// Pop removes and returns the element from the back of the buffer
func (cb *CircularBuffer[T]) Pop() (T, bool) {
	var zeroValue T
	if cb.count == 0 {
		return zeroValue, false
	}
	value := cb.buffer[cb.head]
	cb.buffer[cb.head] = zeroValue
	cb.head = (cb.head + 1) % cb.size
	cb.count--
	return value, true
}

// IsEmpty checks if the buffer is empty
func (cb *CircularBuffer[T]) IsEmpty() bool {
	return cb.count == 0
}

// IsFull checks if the buffer is full
func (cb *CircularBuffer[T]) IsFull() bool {
	return cb.count == cb.size
}
