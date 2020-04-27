package util

import "fmt"

type MaxHeap interface {
	Insert(int) error
	Remove() int
}

type Node struct {
	Val int
	Id  string
}

type maxheap struct {
	heapArray []Node
	size      int
	maxsize   int
}

func NewMaxHeap(maxsize int) *maxheap {
	maxheap := &maxheap{
		heapArray: []Node{},
		size:      0,
		maxsize:   maxsize,
	}
	return maxheap
}

func (m *maxheap) leaf(index int) bool {
	if index >= (m.size/2) && index <= m.size {
		return true
	}
	return false
}

func (m *maxheap) parent(index int) int {
	return (index - 1) / 2
}

func (m *maxheap) leftchild(index int) int {
	return 2*index + 1
}

func (m *maxheap) rightchild(index int) int {
	return 2*index + 2
}

func (m *maxheap) Insert(item Node) error {
	if m.size >= m.maxsize {
		return fmt.Errorf("Heal is ful")
	}
	m.heapArray = append(m.heapArray, item)
	m.size++
	m.downHeapify(m.size - 1)
	return nil
}

func (m *maxheap) swap(first, second int) {
	temp := m.heapArray[first]
	m.heapArray[first] = m.heapArray[second]
	m.heapArray[second] = temp
}

func (m *maxheap) downHeapify(current int) {
	if m.leaf(current) {
		return
	}
	largest := current
	leftChildIndex := m.leftchild(current)
	rightRightIndex := m.rightchild(current)
	//If current is smallest then return
	if leftChildIndex < m.size && m.heapArray[leftChildIndex].Val > m.heapArray[largest].Val {
		largest = leftChildIndex
	}
	if rightRightIndex < m.size && m.heapArray[rightRightIndex].Val > m.heapArray[largest].Val {
		largest = rightRightIndex
	}
	if largest != current {
		m.swap(current, largest)
		m.downHeapify(largest)
	}
	return
}

func (m *maxheap) Remove() Node {
	top := m.heapArray[0]
	m.heapArray[0] = m.heapArray[m.size-1]
	m.heapArray = m.heapArray[:(m.size)-1]
	m.size--
	m.downHeapify(0)
	return top
}
