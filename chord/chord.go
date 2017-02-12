package chord

import (
	"errors"
	"fmt"
	"io"

	"github.com/masu-mi/gimmick.git/sets/s1"
)

// Node is Chord node.
// accept
// key -> id -> Node [ id -> Client ]
//           -> Node [ id -> Client ]
//           -> Node [ id -> Client ]
//           -> Node [ id -> Server ]
// reject
// key -> Node [ key -> id -> ( key -> Client) ]
//     -> Node [ key -> id -> ( key -> Client) ]
//     -> Node [ key -> id -> ( key -> Client) ]
//     -> Node [ key -> id -> Server ]
type Node struct {
	addr string
	id   uint64

	hash func(string) uint64

	successors  []*Node
	predecessor *Node

	finger     []*Node
	nextFinger uint32
	lastIndex  uint32
	failed     bool
}

// func (n *Node) Start(addrs ...string) error {
// 	sock, err := zmq.NewSocket(zmq.PAIR)
// 	if err != nil {
// 		return err
// 	}
// 	s := &http.Server{
// 		Addr:    n.addr,
// 		Handler: http.FileServer(http.Dir(fmt.Sprintf("/tmp/%s/", addr))),
// 	}
// 	s.ListenAndServe()
// }

var (
	ErrEmptyNode = errors.New("node nil")
)

type StorageService interface {
	Put(key string, value io.Reader) error
	Get(key string) (value io.ReadCloser, err error)
}
type proxy struct {
	Client *Client
}

// func (p *proxy) Put(key string, value io.Reader) error {
// 	return p.Client.Put(key, value)
// }
// func (p *proxy) Get(key string) (value io.ReadCloser, err error) {
// 	return p.Client.Get(key, value)
// }

// LocateSuccessor returns client.
// func (n *Node) LocateSuccessor(k uint64) (*Node, error) {
// 	if n == nil || len(n.successors) == 0 {
// 		return nil, ErrEmptyNode
// 	}
// 	if s1.Equal(n.id, k) {
// 		return n, nil
// 	}
// }

// location, routing
func (n *Node) locateSuccessor(k uint64) *Node {
	if n == nil || len(n.successors) == 0 {
		return nil
	}
	if s1.Equal(n.id, k) {
		return n
	}
	// RotationNumber() == 1 likes k <- [n, successor]; close and close
	if s1.RotationNumber(n.id, k, n.successors[0].id) == 1 {
		return n.successors[0]
	}
	next := n.closestPrecedingNode(k)
	return next.locateSuccessor(k)
}
func (n *Node) closestPrecedingNode(k uint64) *Node {
	if n == nil {
		return nil
	}
	if s1.Equal(n.id, k) {
		return n
	}
	if len(n.finger) == 0 && len(n.successors) > 0 {
		return n.successors[0]
	}
	var result *Node
	for i := len(n.finger); i > 0; i-- {
		result = n.finger[i-1]
		if result == nil {
			continue
		}
		if !s1.Equal(n.id, result.id) && s1.RotationNumber(n.id, result.id, k) == 1 {
			break
		}
	}
	return result
}

// Hash changes input key string to id in logical key space.
func (n *Node) Hash(k string) uint64 {
	if n == nil || n.hash == nil {
		return hash(k)
	}
	return n.hash(k)
}
func hash(k string) uint64 {
	return uint64(len(k))
}

// NewNode creates empty Node.
func NewNode(addr string, last uint32, hash func(string) uint64) *Node {
	n := &Node{
		addr: addr, hash: hash,
		lastIndex: last,
		finger:    make([]*Node, 0, last+1),
	}
	n.id = n.Hash(addr)
	return n
}

// Rendezvous
func (n *Node) createNewRing() {
	n.predecessor = nil
	n.successors = []*Node{n}
}
func (n *Node) joinRing(j *Node) {
	n.predecessor = nil
	suc := j.locateSuccessor(n.id)
	n.successors = []*Node{suc}
	// TODO stabilizeは自動実行だけど joinRing直後は即時にする?
	// この手の下位操作と自動的にリングを維持するAPIと層を分けるといい事あるか??
	n.stabilize()
}

// executed periodically to verify and inform successor
func (n *Node) stabilize() {
	prev := n.successors[0].predecessor
	if prev == n {
		return
	}
	if prev != nil && s1.RotationNumber(n.id, prev.id, n.successors[0].id) == 1 {
		n.successors = []*Node{prev}
	}
	n.successors[0].notify(n)
}

// j believes it is predecessor of i
func (n *Node) notify(j *Node) {
	if n.predecessor == nil || s1.RotationNumber(n.predecessor.id, j.id, n.id) == 1 {
		// TODO mentenace service condition
		// transfer keys in the range [j,i) to i; // 要検討
		n.predecessor = j
	}
}

// fixFingers executed periodically to pudate the finger table(n.finger)
func (n *Node) fixFigures() {
	n.nextFinger++
	if n.nextFinger > n.lastIndex {
		n.nextFinger = 0
	}
	terminal := n.locateSuccessor(n.Hash(fmt.Sprintf("%x", n.id+2<<(n.nextFinger))))
	if terminal == n {
		return
	}
	if len(n.finger) < int(n.lastIndex)+1 {
		n.finger = append(n.finger, terminal)
	} else {
		n.finger[n.nextFinger] = terminal
	}
}

// checkPredecessor executed periodically to verify whether predecessor still exists.
func (n *Node) checkPredecessor() {
	if n.predecessor.fail() {
		n.predecessor = nil
	}
}
func (n *Node) checkSuccessors() {
}

// fail check network, Node, host, hardware failer exists.
func (n *Node) fail() bool {
	// for test always OK(false)
	return n.failed
}
