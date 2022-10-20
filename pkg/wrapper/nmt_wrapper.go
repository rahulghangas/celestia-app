package wrapper

import (
	"fmt"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/nmt"
	"github.com/celestiaorg/rsmt2d"
)

// Fulfills the rsmt2d.Tree interface and rsmt2d.TreeConstructorFn function
var (
	_ rsmt2d.TreeConstructorFn = ErasuredNamespacedMerkleTree{}.Constructor
	_ rsmt2d.Tree              = &ErasuredNamespacedMerkleTree{}
)

// ErasuredNamespacedMerkleTree wraps NamespaceMerkleTree to conform to the
// rsmt2d.Tree interface while also providing the correct namespaces to the
// underlying NamespaceMerkleTree. It does this by adding the already included
// namespace to the first half of the tree, and then uses the parity namespace
// ID for each share pushed to the second half of the tree. This allows for the
// namespaces to be included in the erasure data, while also keeping the nmt
// library sufficiently general
type ErasuredNamespacedMerkleTree struct {
	squareSize uint64 // note: this refers to the width of the original square before erasure-coded
	options    []nmt.Option
	tree       *nmt.NamespacedMerkleTree
}

// NewErasuredNamespacedMerkleTree issues a new ErasuredNamespacedMerkleTree. squareSize must be greater than zero
func NewErasuredNamespacedMerkleTree(squareSize uint64, setters ...nmt.Option) ErasuredNamespacedMerkleTree {
	if squareSize == 0 {
		panic("cannot create a ErasuredNamespacedMerkleTree of squareSize == 0")
	}
	tree := nmt.New(appconsts.NewBaseHashFunc(), setters...)
	return ErasuredNamespacedMerkleTree{squareSize: squareSize, options: setters, tree: tree}
}

// Constructor acts as the rsmt2d.TreeConstructorFn for
// ErasuredNamespacedMerkleTree
func (w ErasuredNamespacedMerkleTree) Constructor() rsmt2d.Tree {
	newTree := NewErasuredNamespacedMerkleTree(w.squareSize, w.options...)
	return &newTree
}

// Push adds the provided data to the underlying NamespaceMerkleTree, and
// automatically uses the first DefaultNamespaceIDLen number of bytes as the
// namespace unless the data pushed to the second half of the tree. Fulfills the
// rsmt.Tree interface. NOTE: panics if an error is encountered while pushing or
// if the tree size is exceeded.
func (w *ErasuredNamespacedMerkleTree) Push(data []byte, idx rsmt2d.SquareIndex) {
	if idx.Axis+1 > 2*uint(w.squareSize) || idx.Cell+1 > 2*uint(w.squareSize) {
		panic(fmt.Sprintf("pushed past predetermined square size: boundary at %d index at %+v", 2*w.squareSize, idx))
	}
	nidAndData := make([]byte, appconsts.NamespaceSize+len(data))
	copy(nidAndData[appconsts.NamespaceSize:], data)
	// use the parity namespace if the cell is not in Q0 of the extended data square
	if idx.Axis+1 > uint(w.squareSize) || idx.Cell+1 > uint(w.squareSize) {
		copy(nidAndData[:appconsts.NamespaceSize], appconsts.ParitySharesNamespaceID)
	} else {
		copy(nidAndData[:appconsts.NamespaceSize], data[:appconsts.NamespaceSize])
	}
	// push to the underlying tree
	err := w.tree.Push(nidAndData)
	// panic on error
	if err != nil {
		panic(err)
	}
}

// Root fulfills the rsmt.Tree interface by generating and returning the
// underlying NamespaceMerkleTree Root.
func (w *ErasuredNamespacedMerkleTree) Root() []byte {
	return w.tree.Root()
}

func (w *ErasuredNamespacedMerkleTree) Prove(ind int) (nmt.Proof, error) {
	return w.tree.Prove(ind)
}

// Tree returns the underlying NamespacedMerkleTree
func (w *ErasuredNamespacedMerkleTree) Tree() *nmt.NamespacedMerkleTree {
	return w.tree
}
