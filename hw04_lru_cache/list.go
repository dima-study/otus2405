package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

var _ List = (*list)(nil)

type list struct {
	len   int       // len holds number of items in the list.
	front *ListItem // front holds pointer to the first element in the list.
	back  *ListItem // back holds pointer to the last element in the list.
}

// NewList creates new List.
func NewList() List {
	return &list{}
}

// Len returns list length.
func (l *list) Len() int {
	return l.len
}

// Front returns the first item in the list.
func (l *list) Front() *ListItem {
	return l.front
}

// Back returns the last item in the list.
func (l *list) Back() *ListItem {
	return l.back
}

// PushFront adds value to the front of the list and returns added item.
func (l *list) PushFront(val interface{}) *ListItem {
	front := &ListItem{
		Value: val,
	}

	l.pushItemFront(front)

	return front
}

// pushItemFront adds item to the front of the list.
// Is being used internally by PushFront and MoveToFront.
func (l *list) pushItemFront(item *ListItem) {
	if l.front == nil {
		// The list front is nil, so there are no items in the list at all
		// and the new item becomes the back.
		l.back = item
	} else {
		l.front.Prev = item
		item.Next = l.front
	}

	l.front = item
	l.len++
}

// PushBack adds value to the back of the list and returns added item.
func (l *list) PushBack(val interface{}) *ListItem {
	back := &ListItem{
		Value: val,
	}

	if l.back == nil {
		// The list back is nil, so there are no items in the list at all
		// and the new item becomes the front.
		l.front = back
	} else {
		l.back.Next = back
		back.Prev = l.back
	}

	l.back = back
	l.len++

	return back
}

// Remove removes item from the list.
func (l *list) Remove(i *ListItem) {
	prev := i.Prev
	next := i.Next

	// i is the front of the list, so the next item to i becomes the new front.
	if prev == nil {
		l.front = next
		next.Prev = nil
	}

	// i is the back of the list, so the item before i becomes the new back.
	if next == nil {
		l.back = prev
		prev.Next = nil
	}

	// i is in the middle, link prev-to-next (and vise-verse).
	if prev != nil && next != nil {
		prev.Next, next.Prev = next, prev
	}

	l.len--

	i.Prev = nil
	i.Next = nil
}

// MoveToFront moves item to the list head.
func (l *list) MoveToFront(i *ListItem) {
	l.Remove(i)
	l.pushItemFront(i)
}
