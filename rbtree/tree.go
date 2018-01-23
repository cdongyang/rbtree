package rbtree

import (
	"errors"
	"unsafe"
)

var (
	ErrNotInTree  = errors.New("Iterator is not a node of this tree")
	ErrNoLast     = errors.New("begin of tree has no Last()")
	ErrNoNext     = errors.New("end of tree has no Next()")
	ErrEraseEmpty = errors.New("can't erase empty node")
)

// Iterator is the interface of Node
type Iterator interface {
	Next() Iterator
	Last() Iterator
	GetData() (data interface{})
	GetKey() (key interface{})
	GetValue() (value interface{})
	SetValue(value interface{})
	CopyData(src Iterator)
	GetTree() Treer
}

type colorType bool

const (
	red   = false
	black = true
)

// 将value赋值给interface时编译器取这个value的runtime type和在堆上分配的value对象的指针绑定成一个iface/eface struct
// 将pointer赋值给interface时编译器取这个pointer的指向的值的runtime type并和这个pointer绑定成一个iface/eface struct
type eface struct {
	_type   unsafe.Pointer
	pointer unsafe.Pointer
}

// inherit this node type to set the tree node privite
type _node struct {
	Node
}

// Node is the node of Tree, it implement Iterator
// inherit Node should not use pointer
// because the operation of Tree treate it as value
// also, pointer will increase the GC pressure
type Node struct {
	child  [2]unsafe.Pointer
	parent unsafe.Pointer
	tree   *Tree
	color  colorType
}

// Next return next Iterator of this
func (node *Node) Next() Iterator {
	return node.GetTree().Next(node)
}

// Last return last Iterator of this
func (node *Node) Last() Iterator {
	return node.GetTree().Last(node)
}

// GetData get the data of this
func (node *Node) GetData() interface{} {
	// do nothing here, just implement the interface
	return nil
}

// GetKey get the compare key of this
func (node *Node) GetKey() interface{} {
	// do nothing here, just implement the interface
	return nil
}

// GetValue get the value of this
func (node *Node) GetValue() interface{} {
	// do nothing here, just implement the interface
	return nil
}

// SetValue set the value of this
func (node *Node) SetValue(interface{}) {
	// do nothing here, just implement the interface
}

//CopyData copy the node data to this from src
func (node *Node) CopyData(src Iterator) {
	// do nothing here, just implement the interface
}

// GetTree get the Treer of this
func (node *Node) GetTree() Treer {
	return (*Tree)(node.tree).tree
}

// Treer is the interface of Tree
type Treer interface {
	SameIterator(a, b Iterator) bool
	Compare(key1, key2 unsafe.Pointer) int
	Clear()
	Unique() bool
	Size() int
	Empty() bool
	Begin() Iterator
	End() Iterator
	Last(node Iterator) Iterator
	Next(node Iterator) Iterator
	Count(key interface{}) int
	EqualRange(key interface{}) (beg Iterator, end Iterator)
	Find(key interface{}) Iterator
	Insert(data interface{}) (Iterator, bool)
	Erase(key interface{}) int
	EraseIterator(node Iterator)
	EraseIteratorRange(beg Iterator, end Iterator) int
	LowerBound(key interface{}) Iterator
	UpperBound(key interface{}) Iterator

	init(
		tree Treer,
		header Iterator,
		nodeOffset uintptr,
		newNode func(interface{}) Iterator,
		deleteNode func(Iterator),
		compare func(unsafe.Pointer, unsafe.Pointer) int,
		getKeyPointer func(unsafe.Pointer) unsafe.Pointer,
		unique bool,
	)
}

// inherit this tree type to set tree privite
type _tree struct {
	Tree
}

// Tree is a red-black tree
type Tree struct {
	// header is a virtual node, it represent the tree's end
	header unsafe.Pointer
	// nodeType is the runtime type of inherit Node type
	nodeType unsafe.Pointer
	// tree represent the interface of inherit Tree type
	tree Treer
	// size represent the size of tree
	size int
	// newNode is the func to create a new node of inherit Node type
	newNode func(interface{}) Iterator
	// deleteNode is the func to delete a node of inherit Node type
	// return to a pool or just do nothing and let it recycle by GC
	deleteNode func(Iterator)
	// compare is the compare func to compare two node key
	compare       func(a, b unsafe.Pointer) int
	getKeyPointer func(unsafe.Pointer) unsafe.Pointer
	// represent the offset of Node in the inherit Node type
	nodeOffset uintptr
	// unique is to mark whether the tree node is unique
	unique bool
}

// NewTreer create a new tree with it's element
func NewTreer(
	t Treer,
	header Iterator,
	nodeOffset uintptr,
	newNode func(interface{}) Iterator,
	deleteNode func(Iterator),
	compare func(unsafe.Pointer, unsafe.Pointer) int,
	getKeyPointer func(unsafe.Pointer) unsafe.Pointer,
	unique bool) Treer {
	t.init(t, header, nodeOffset, newNode, deleteNode, compare,
		getKeyPointer,
		unique)
	return t
}

func (t *Tree) init(
	tree Treer,
	header Iterator,
	nodeOffset uintptr,
	newNode func(interface{}) Iterator,
	deleteNode func(Iterator),
	compare func(unsafe.Pointer, unsafe.Pointer) int,
	getKeyPointer func(unsafe.Pointer) unsafe.Pointer,
	unique bool) {
	t.nodeOffset = nodeOffset
	t.nodeType = iterator2type(header)
	t.tree = tree
	t.header = iterator2pointer(header)
	t.setTree(t.header, interface2pointer(tree))
	t.setColor(t.header, red)
	*t.mostPoiter(0) = t.end()
	*t.mostPoiter(1) = t.end()
	*t.rootPoiter() = t.end()
	t.size = 0
	t.newNode = func(data interface{}) Iterator {
		var node = newNode(data)
		var nodePointer = iterator2pointer(node)
		t.setChild(nodePointer, 0, t.end())
		t.setChild(nodePointer, 1, t.end())
		t.setParent(nodePointer, t.end())
		t.setTree(nodePointer, interface2pointer(tree))
		t.setColor(nodePointer, red)
		return node
	}
	t.deleteNode = func(node Iterator) {
		var nodePointer = iterator2pointer(node)
		t.setChild(nodePointer, 0, nil)
		t.setChild(nodePointer, 1, nil)
		t.setParent(nodePointer, nil)
		t.setTree(nodePointer, nil)
		deleteNode(node)
	}
	t.compare = compare
	t.getKeyPointer = getKeyPointer
	t.unique = unique
}

func unsafeSameIterator(a, b Iterator) bool {
	return (*eface)(unsafe.Pointer(&a)).pointer == (*eface)(unsafe.Pointer(&b)).pointer
}

func sameNode(a, b unsafe.Pointer) bool {
	return a == b
}

func (t *Tree) SameIterator(a, b Iterator) bool {
	return unsafeSameIterator(a, b)
}

func (t *Tree) Compare(key1, key2 unsafe.Pointer) int {
	return t.compare(key1, key2)
}

func (t *Tree) Clear() {
	t.clear(t.root())
	*t.mostPoiter(0) = t.end()
	*t.mostPoiter(1) = t.end()
	*t.rootPoiter() = t.end()
	t.size = 0
}

func (t *Tree) clear(root unsafe.Pointer) {
	if sameNode(root, t.end()) {
		return
	}
	t.clear(t.getChild(root, 0))
	t.clear(t.getChild(root, 1))
	t.deleteNode(t.pointer2iterator(root))
}

// Unique return a bool value represent whether the tree node is unique
func (t *Tree) Unique() bool {
	return t.unique
}

// Size return the size of tree, which represent the number of node in tree
func (t *Tree) Size() int {
	return t.size
}

// Empty return a bool value represent whether the tree has no node
func (t *Tree) Empty() bool {
	return t.size == 0
}

// Begin return the Iterator of first node
// if the tree is empty, it will equal to End
func (t *Tree) Begin() Iterator {
	return t.pointer2iterator(t.begin())
}

func (t *Tree) begin() unsafe.Pointer {
	return t.most(0)
}

// End return the End of tree, but it's not a tree node of tree
// just like a[10] of var a [10]int, it's the pointer to end
func (t *Tree) End() Iterator {
	return t.pointer2iterator(t.end())
}

func (t *Tree) end() unsafe.Pointer {
	return t.header
}

// Next return the next Iterator of node in this tree
// if node has no next Iterator, it will panic
func (t *Tree) Next(node Iterator) Iterator {
	if !t.sameTree(node) {
		panic(ErrNotInTree)
	}
	return t.pointer2iterator(t.next(iterator2pointer(node)))
}
func (t *Tree) next(node unsafe.Pointer) unsafe.Pointer {
	if sameNode(node, t.end()) {
		panic(ErrNoNext)
	}
	if sameNode(node, t.most(1)) {
		return t.end()
	}
	return t.gothrough(1, node)
}

// Last return the last Iterator of node in this tree
// if node has no last Iterator, it will panic
func (t *Tree) Last(node Iterator) Iterator {
	if !t.sameTree(node) {
		panic(ErrNotInTree)
	}
	return t.pointer2iterator(t.last(iterator2pointer(node)))
}
func (t *Tree) last(node unsafe.Pointer) unsafe.Pointer {
	if sameNode(node, t.begin()) {
		panic(ErrNoLast)
	}
	if sameNode(node, t.end()) {
		return t.most(1)
	}
	return t.gothrough(0, node)
}

func (t *Tree) gothrough(ch int, node unsafe.Pointer) unsafe.Pointer {
	if !sameNode(t.getChild(node, ch), t.end()) {
		node = t.getChild(node, ch)
		for !sameNode(t.getChild(node, ch^1), t.end()) {
			node = t.getChild(node, ch^1)
		}
		return node
	}
	for !sameNode(t.getParent(node), t.end()) && sameNode(t.getChild(t.getParent(node), ch), node) {
		node = t.getParent(node)
	}
	return t.getParent(node)
}

// Count return the num of node key equal to key in this tree
func (t *Tree) Count(key interface{}) (count int) {
	var keyPointer = interface2pointer(key)
	if t.unique {
		if sameNode(t.find(keyPointer), t.end()) {
			return 0
		}
		return 1
	}
	var beg = t.lowerBound(keyPointer)
	for !sameNode(beg, t.end()) && t.compare(t.getKeyPointer(beg), keyPointer) == 0 {
		beg = t.next(beg)
		count++
	}
	return count
}

// EqualRange return the Iterator range of equal key node in this tree
func (t *Tree) EqualRange(key interface{}) (beg, end Iterator) {
	return t.LowerBound(key), t.UpperBound(key)
}

// Find return the Iterator of key in this tree
// if the key is not exist in this tree, result will be the End of tree
// if there has multi node key equal to key, result will be random one
func (t *Tree) Find(key interface{}) Iterator {
	return t.pointer2iterator(t.find(noescape(interface2pointer(key))))
}
func (t *Tree) find(keyPointer unsafe.Pointer) unsafe.Pointer {
	var root = t.root()
	for {
		if sameNode(root, t.end()) {
			return root
		}
		switch cmp := t.compare(keyPointer, t.getKeyPointer(root)); {
		case cmp == 0:
			return root
		case cmp < 0:
			root = t.getChild(root, 0)
		case cmp > 0:
			root = t.getChild(root, 1)
		}
	}
}

// Insert insert a new node with data to tree
// it return the insert node Iterator and true when success insert
// otherwise, it return the end of tree and false
func (t *Tree) Insert(data interface{}) (Iterator, bool) {
	iter, ok := t.insert(data, interface2pointer(data))
	return t.pointer2iterator(iter), ok
}
func (t *Tree) insert(data interface{}, key unsafe.Pointer) (unsafe.Pointer, bool) {
	var root = t.root()
	var rootPoiter = t.rootPoiter()
	if sameNode(root, t.end()) {
		t.size++
		*rootPoiter = iterator2pointer(t.newNode(data))
		t.insertAdjust(*rootPoiter)
		*t.mostPoiter(0) = *rootPoiter
		*t.mostPoiter(1) = *rootPoiter
		return *rootPoiter, true
	}
	var parent = t.getParent(root)
	for !sameNode(root, t.end()) {
		parent = root
		switch cmp := t.compare(key, t.getKeyPointer(root)); {
		case cmp == 0:
			if t.unique {
				return t.end(), false
			}
			fallthrough
		case cmp < 0:
			rootPoiter = t.getChildPointer(root, 0)
			root = *rootPoiter
		case cmp > 0:
			rootPoiter = t.getChildPointer(root, 1)
			root = *rootPoiter
		}
	}
	t.size++
	*rootPoiter = iterator2pointer(t.newNode(data))
	t.setParent((*rootPoiter), parent)
	for ch := 0; ch < 2; ch++ {
		if sameNode(parent, t.most(ch)) && sameNode(t.getChild(parent, ch), *rootPoiter) {
			*t.mostPoiter(ch) = *rootPoiter
		}
	}
	t.insertAdjust(*rootPoiter)
	return *rootPoiter, true
}

//insert node is default red
func (t *Tree) insertAdjust(node unsafe.Pointer) {
	var parent = t.getParent(node)
	if sameNode(parent, t.end()) {
		//fmt.Println("case 1: insert")
		//node is root,set black
		t.setColor(node, black)
		return
	}
	if t.getColor(parent) == black {
		//fmt.Println("case 2: insert")
		//if parent is black,do nothing
		return
	}

	//parent is red,grandpa can't be empty and color is black
	var grandpa = t.getParent(parent)
	var parentCh = 0
	if sameNode(t.getChild(grandpa, 1), parent) {
		parentCh = 1
	}

	var uncle = t.getChild(grandpa, parentCh^1)
	if !sameNode(uncle, t.end()) && t.getColor(uncle) == red {
		//fmt.Println("case 3: insert")
		//uncle is red
		t.setColor(parent, black)
		t.setColor(grandpa, red)
		t.setColor(uncle, black)
		t.insertAdjust(grandpa)
		return
	}

	var childCh = 0
	if sameNode(t.getChild(parent, 1), node) {
		childCh = 1
	}
	if childCh != parentCh {
		//fmt.Println("case 4: insert")
		t.rotate(parentCh, node)
		var tmp = parent
		parent = node
		node = tmp
	}

	//fmt.Println("case 5: insert")
	t.setColor(parent, black)
	t.setColor(grandpa, red)
	t.rotate(parentCh^1, parent)
}

// Erase erase all the node keys equal to key in this tree and return the number of erase node
func (t *Tree) Erase(key interface{}) (count int) {
	var keyPointer = noescape(interface2pointer(key))
	if t.unique {
		var iter = t.find(keyPointer)
		if sameNode(iter, t.end()) {
			return 0
		}
		t.eraseIterator(iter)
		return 1
	}
	var beg = t.lowerBound(keyPointer)
	for !sameNode(beg, t.end()) && t.compare(keyPointer, t.getKeyPointer(beg)) == 0 {
		var tmp = t.next(beg)
		t.eraseIterator(beg)
		beg = tmp
		count++
	}
	return count
}

// EraseIterator erase node from the tree
// if node is not in tree, it will panic
func (t *Tree) EraseIterator(node Iterator) {
	if !t.sameTree(node) {
		panic(ErrNotInTree)
	}
	t.eraseIterator(iterator2pointer(node))
}
func (t *Tree) eraseIterator(node unsafe.Pointer) {
	if sameNode(node, t.end()) {
		panic(ErrEraseEmpty)
	}
	t.size--
	if !sameNode(t.getChild(node, 0), t.end()) && !sameNode(t.getChild(node, 1), t.end()) {
		//if node has two child,it's last node must has no more than one child,copy to node and erase last node
		var tmp = t.last(node)
		t.pointer2iterator(node).CopyData(t.pointer2iterator(tmp))
		node = tmp
	}
	//adjust leftmost and rightmost
	for ch := 0; ch < 2; ch++ {
		if sameNode(t.most(ch), node) {
			if ch == 0 {
				*t.mostPoiter(ch) = t.next(node)
			} else {
				*t.mostPoiter(ch) = t.last(node)
			}
		}
	}
	var child = t.end()
	if !sameNode(t.getChild(node, 0), t.end()) {
		child = t.getChild(node, 0)
	} else if !sameNode(t.getChild(node, 1), t.end()) {
		child = t.getChild(node, 1)
	}
	var parent = t.getParent(node)
	if !sameNode(child, t.end()) {
		t.setParent(child, parent)
	}
	if sameNode(parent, t.end()) {
		*t.rootPoiter() = child
	} else if sameNode(t.getChild(parent, 0), node) {
		t.setChild(parent, 0, child)
	} else {
		t.setChild(parent, 1, child)
	}
	if t.getColor(node) == black { //if node is red,just erase,otherwise adjust
		t.eraseAdjust(child, parent)
		//fmt.Println("eraseAdjust:")
	}
	t.deleteNode(t.pointer2iterator(node))
	return
}

func (t *Tree) eraseAdjust(node, parent unsafe.Pointer) {
	if sameNode(parent, t.end()) {
		//node is root
		//fmt.Println("case 1: erase")
		if !sameNode(node, t.end()) {
			t.setColor(node, black)
		}
		return
	}
	if t.mustGetColor(node) == red {
		//node is red,just set black
		//fmt.Println("case 2: erase")
		t.setColor(node, black)
		return
	}
	var nodeCh = 0
	if sameNode(t.getChild(parent, 1), node) {
		nodeCh = 1
	}
	var brother = t.getChild(parent, nodeCh^1)
	//after case 1 parent must not be empty node and after case 2 node must be black
	if t.getColor(parent) == red {
		//parent is red,brother must be black but can't be empty node,because the path has a black node more
		if t.mustGetColor(t.getChild(brother, 0)) == black && t.mustGetColor(t.getChild(brother, 1)) == black {
			//fmt.Println("case 3: erase")
			t.setColor(brother, red)
			t.setColor(parent, black)
			return
		}
		if !sameNode(brother, t.end()) && t.mustGetColor(t.getChild(brother, nodeCh)) == red {
			//fmt.Println("case 4: erase", nodeCh)
			t.setColor(parent, black)
			t.rotate(nodeCh^1, t.getChild(brother, nodeCh))
			t.rotate(nodeCh, t.getChild(parent, nodeCh^1))
			return
		}
		//fmt.Println("case 5: erase")
		t.rotate(nodeCh, brother)
		return
	}
	//parent is black
	if t.mustGetColor(brother) == red {
		//brother is red, it's children must be black
		//fmt.Println("case 6: erase")
		t.setColor(brother, black)
		t.setColor(parent, red)
		t.rotate(nodeCh, brother)
		t.eraseAdjust(node, parent) //goto redParent then end
		return
	}
	//brother is black
	if t.mustGetColor(t.getChild(brother, 0)) == black && t.mustGetColor(t.getChild(brother, 1)) == black {
		//fmt.Println("case 7: erase")
		t.setColor(brother, red)
		t.eraseAdjust(parent, t.getParent(parent))
		return
	}
	if t.mustGetColor(t.getChild(brother, nodeCh)) == red {
		//fmt.Println("case 8: erase", nodeCh)
		t.setColor(t.getChild(brother, nodeCh), black)
		t.rotate(nodeCh^1, t.getChild(brother, nodeCh))
		t.rotate(nodeCh, t.getChild(parent, nodeCh^1))
		return
	}
	//fmt.Println("case 9: erase", nodeCh)
	t.setColor(t.getChild(brother, nodeCh^1), black)
	t.rotate(nodeCh, brother)
}

// EraseIteratorRange erase the given iterator range
// if the given range is not in this tree, it will panic with ErrNoInTree
// if end can get beg after multi Next method, it will panic with ErrNoLast
func (t *Tree) EraseIteratorRange(beg, end Iterator) (count int) {
	return t.eraseIteratorRange(iterator2pointer(beg), iterator2pointer(end))
}
func (t *Tree) eraseIteratorRange(beg, end unsafe.Pointer) (count int) {
	for !sameNode(beg, end) {
		var tmp = t.next(beg)
		t.eraseIterator(beg)
		beg = tmp
		count++
	}
	return count
}

// LowerBound return the first Iterator greater than or equal to key
func (t *Tree) LowerBound(key interface{}) Iterator {
	return t.pointer2iterator(t.lowerBound(noescape(interface2pointer(key))))
}
func (t *Tree) lowerBound(keyPointer unsafe.Pointer) unsafe.Pointer {
	var root = t.root()
	var parent = t.end()
	for {
		if root == t.end() {
			if sameNode(parent, t.end()) {
				return parent
			} else if t.compare(keyPointer, t.getKeyPointer(parent)) <= 0 {
				return parent
			}
			return t.next(parent)
		}
		parent = root
		if t.compare(keyPointer, t.getKeyPointer(root)) > 0 {
			root = t.getChild(root, 1)
		} else {
			root = t.getChild(root, 0)
		}
	}
}

// UpperBound return the first Iterator greater than key
func (t *Tree) UpperBound(key interface{}) Iterator {
	return t.pointer2iterator(t.upperBound(noescape(interface2pointer(key))))
}
func (t *Tree) upperBound(keyPointer unsafe.Pointer) unsafe.Pointer {
	var root = t.root()
	var parent = t.end()
	for {
		if root == t.end() {
			if sameNode(parent, t.end()) {
				return parent
			} else if t.compare(keyPointer, t.getKeyPointer(parent)) < 0 {
				return parent
			}
			return t.next(parent)
		}
		parent = root
		if t.compare(keyPointer, t.getKeyPointer(root)) >= 0 {
			root = t.getChild(root, 1)
		} else {
			root = t.getChild(root, 0)
		}
	}
}

//ch = 0:take node for center,left rotate parent down,node is parent's right child
//ch = 1:take node for center,right rotate parent down,node is parent's left child
func (t *Tree) rotate(ch int, node unsafe.Pointer) {
	var (
		tmp     = t.getChild(node, ch)
		parent  = t.getParent(node)
		grandpa = t.getParent(parent)
	)
	t.setChild(node, ch, parent)
	t.setChild(parent, ch^1, tmp)

	if !sameNode(tmp, t.end()) {
		t.setParent(tmp, parent)
	}
	t.setParent(parent, node)
	t.setParent(node, grandpa)
	if sameNode(grandpa, t.end()) {
		*t.rootPoiter() = node
		return
	}
	if sameNode(t.getChild(grandpa, 0), parent) {
		t.setChild(grandpa, 0, node)
	} else {
		t.setChild(grandpa, 1, node)
	}
}

func (t *Tree) root() unsafe.Pointer {
	return t.getParent(t.header)
}

func (t *Tree) rootPoiter() *unsafe.Pointer {
	return t.getParentPointer(t.header)
}

//ch = 0: leftmost; ch = 1: rightmost
func (t *Tree) most(ch int) unsafe.Pointer {
	return t.getChild(t.header, ch)
}

//ch = 0: leftmostPoiter; ch = 1: rightmostPoiter
func (t *Tree) mostPoiter(ch int) *unsafe.Pointer {
	return t.getChildPointer(t.header, ch)
}

func (t *Tree) mustGetColor(node unsafe.Pointer) colorType {
	if !sameNode(node, t.end()) {
		return t.getColor(node)
	}
	return black
}

func (t *Tree) sameTree(node Iterator) bool {
	return unsafe.Pointer(t) == interface2pointer(node.GetTree())
}

func (t *Tree) pointer2iterator(node unsafe.Pointer) Iterator {
	var tmp = [2]unsafe.Pointer{t.nodeType, node}
	return *(*Iterator)(unsafe.Pointer(&tmp))
}

func (t *Tree) getNode(node unsafe.Pointer) *Node {
	return (*Node)(unsafe.Pointer(uintptr(node) + t.nodeOffset))
}

func (t *Tree) getChild(node unsafe.Pointer, ch int) unsafe.Pointer {
	return *getNodePointer(node, t.nodeOffset+offsetChild[ch])
}

func (t *Tree) getChildPointer(node unsafe.Pointer, ch int) *unsafe.Pointer {
	return getNodePointer(node, t.nodeOffset+offsetChild[ch])
}

func (t *Tree) setChild(node unsafe.Pointer, ch int, child unsafe.Pointer) {
	*getNodePointer(node, t.nodeOffset+offsetChild[ch]) = child
}

func (t *Tree) getParent(node unsafe.Pointer) unsafe.Pointer {
	return *getNodePointer(node, t.nodeOffset+offsetParent)
}

func (t *Tree) getParentPointer(node unsafe.Pointer) *unsafe.Pointer {
	return getNodePointer(node, t.nodeOffset+offsetParent)
}

func (t *Tree) setParent(node unsafe.Pointer, parent unsafe.Pointer) {
	*getNodePointer(node, t.nodeOffset+offsetParent) = parent
}

func (t *Tree) getColor(node unsafe.Pointer) colorType {
	return *getColorPointer(node, t.nodeOffset+offsetColor)
}

func (t *Tree) setColor(node unsafe.Pointer, color colorType) {
	*getColorPointer(node, t.nodeOffset+offsetColor) = color
}

func (t *Tree) setTree(node unsafe.Pointer, tree unsafe.Pointer) {
	*getNodePointer(node, t.nodeOffset+offsetTree) = tree
}

/*func (t *Tree) compare(a, b unsafe.Pointer) int {
	fun := (*func(a, b unsafe.Pointer) int)(t.compareFun)
	return (*fun)(a, b)
}*/

var (
	offsetChild [2]uintptr
	offsetParent,
	offsetTree,
	offsetColor uintptr
)

func init() {
	var node = &Node{}
	offsetChild[0] = uintptr(unsafe.Pointer(&node.child[0])) - uintptr(unsafe.Pointer(node))
	offsetChild[1] = uintptr(unsafe.Pointer(&node.child[1])) - uintptr(unsafe.Pointer(node))
	offsetParent = uintptr(unsafe.Pointer(&node.parent)) - uintptr(unsafe.Pointer(node))
	offsetTree = uintptr(unsafe.Pointer(&node.tree)) - uintptr(unsafe.Pointer(node))
	offsetColor = uintptr(unsafe.Pointer(&node.color)) - uintptr(unsafe.Pointer(node))
	//fmt.Println(offsetChild[0], offsetChild[1], offsetParent, offsetTree, offsetColor)
}

//var GetNodeCount = 0

func getNodePointer(node unsafe.Pointer, offset uintptr) *unsafe.Pointer {
	//GetNodeCount++
	return (*unsafe.Pointer)(unsafe.Pointer(uintptr(node) + offset))
}

func getColorPointer(node unsafe.Pointer, offset uintptr) *colorType {
	return (*colorType)(unsafe.Pointer(uintptr(node) + offset))
}

func iterator2eface(node Iterator) eface {
	return *(*eface)(unsafe.Pointer(&node))
}

// the first pointer is type
func iterator2type(node Iterator) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&node))
}

// this second pointer is pointer
func iterator2pointer(node Iterator) unsafe.Pointer {
	return (*[2]unsafe.Pointer)(unsafe.Pointer(&node))[1]
}

func eface2iterator(node eface) Iterator {
	return *(*Iterator)(unsafe.Pointer(&node))
}

func interface2eface(node interface{}) eface {
	return *(*eface)(unsafe.Pointer(&node))
}

func eface2interface(node eface) interface{} {
	return *(*interface{})(unsafe.Pointer(&node))
}

func interface2type(a interface{}) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&a))
}

func interface2pointer(a interface{}) unsafe.Pointer {
	return (*eface)(unsafe.Pointer(&a)).pointer
}

func CompareInt(a, b unsafe.Pointer) int {
	return *(*int)(a) - *(*int)(b)
}

// copy from package runtime
// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
