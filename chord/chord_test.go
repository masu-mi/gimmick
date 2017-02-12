package chord

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/masu-mi/gimmick.git/sets/s1"
)

func createTestKey(id uint64) string {
	return fmt.Sprintf("%x", id)
}

// TestLocateSuccessor
func TestLocateSuccessor(t *testing.T) {
	type testCase struct {
		inputKey   string
		expectedID uint64
	}
	type test struct {
		title    string
		ring     []*Node
		testCase []testCase
	}
	for _, set := range []test{
		test{
			title: "one node ring(id:0(min))",
			ring:  generateNodes(1, 0, 1),
			testCase: []testCase{
				testCase{inputKey: createTestKey(0), expectedID: 0},
				testCase{inputKey: createTestKey(1), expectedID: 0},
				testCase{inputKey: createTestKey(2), expectedID: 0},
			},
		},
		test{
			title: "one node ring(id:1)",
			ring:  generateNodes(1, 1, 1),
			testCase: []testCase{
				testCase{inputKey: createTestKey(0), expectedID: 1},
				testCase{inputKey: createTestKey(1), expectedID: 1},
				testCase{inputKey: createTestKey(2), expectedID: 1},
			},
		},
		test{
			title: "2 node ring(id:0,4)",
			ring:  generateNodes(2, 0, 4),
			testCase: []testCase{
				testCase{inputKey: createTestKey(0), expectedID: 0},
				testCase{inputKey: createTestKey(1), expectedID: 4},
				testCase{inputKey: createTestKey(2), expectedID: 4},
				testCase{inputKey: createTestKey(3), expectedID: 4},
				testCase{inputKey: createTestKey(4), expectedID: 4},
				testCase{inputKey: createTestKey(5), expectedID: 0},
				testCase{inputKey: createTestKey(6), expectedID: 0},
				testCase{inputKey: createTestKey(7), expectedID: 0},
				testCase{inputKey: createTestKey(8), expectedID: 0},
			},
		},
		test{
			title: "2 node ring(id:1,5)",
			ring:  generateNodes(2, 1, 4),
			testCase: []testCase{
				testCase{inputKey: createTestKey(0), expectedID: 1},
				testCase{inputKey: createTestKey(1), expectedID: 1},
				testCase{inputKey: createTestKey(2), expectedID: 5},
				testCase{inputKey: createTestKey(3), expectedID: 5},
				testCase{inputKey: createTestKey(4), expectedID: 5},
				testCase{inputKey: createTestKey(5), expectedID: 5},
				testCase{inputKey: createTestKey(6), expectedID: 1},
				testCase{inputKey: createTestKey(7), expectedID: 1},
				testCase{inputKey: createTestKey(8), expectedID: 1},
			},
		},
	} {
		setupRingStatically(set.ring, 1)
		assertTestRingStatically(t, set.ring)
		t.Run(set.title, func(t *testing.T) {
			for _, n := range set.ring {
				for _, c := range set.testCase {
					act := n.locateSuccessor(n.Hash(c.inputKey))
					if act == nil {
						t.Errorf("nil was returned with starting from id:%d; expectedID:%d, key:%s",
							n.id, c.expectedID, c.inputKey)
					} else if !s1.Equal(act.id, c.expectedID) {
						t.Errorf("invalid node returned(%+v) with starting from id:%d; expectedID:%d, key:%s",
							act, n.id, c.expectedID, c.inputKey)
					}
				}
			}
		})
	}
}

func TestClosestPrecedingNode(t *testing.T) {
	type testCase struct {
		inputKey   string
		nilReturn  bool
		expectedID uint64
	}
	type test struct {
		cordinator *Node
		finger     []*Node
		cases      []testCase
	}
	for _, tests := range []test{
		test{
			cordinator: &Node{id: 2, hash: testHash}, finger: nil,
			cases: []testCase{
				testCase{inputKey: createTestKey(1), nilReturn: true},
			},
		},
		test{
			cordinator: &Node{id: 2, hash: testHash}, finger: []*Node{},
			cases: []testCase{
				testCase{inputKey: createTestKey(1), nilReturn: true},
			},
		},
		test{
			cordinator: &Node{id: 2, hash: testHash}, finger: []*Node{&Node{id: 10}, &Node{id: 20}},
			cases: []testCase{
				testCase{inputKey: createTestKey(1), expectedID: 20}, // sucessor: 2
				testCase{inputKey: createTestKey(2), expectedID: 2},  // sucessor: 2
				testCase{inputKey: createTestKey(10), expectedID: 10},
				testCase{inputKey: createTestKey(11), expectedID: 10},
				testCase{inputKey: createTestKey(20), expectedID: 20},
				testCase{inputKey: createTestKey(21), expectedID: 20},
			},
		},
	} {
		entrypoint := tests.cordinator
		entrypoint.finger = tests.finger
		var ids []uint64
		for _, n := range entrypoint.finger {
			ids = append(ids, n.id)
		}
		t.Run(fmt.Sprintf("node(id:%d, finger:%+v).closestPrecedingNode($key)", entrypoint.id, ids), func(t *testing.T) {
			for _, c := range tests.cases {
				act := entrypoint.closestPrecedingNode(entrypoint.Hash(c.inputKey))
				if act == nil {
					if !c.nilReturn {
						t.Errorf("nil was returned.(keyid: %d)", len(c.inputKey))
					}
				} else if !s1.Equal(act.id, c.expectedID) {
					t.Errorf("key: %s, result(%d) unmatch expected(%d)",
						c.inputKey, act.id, c.expectedID,
					)
				}
			}
		})
	}
	// closestPrecedingNode is safe with that finger's len == 0 or finger is nil.
	// node which id is max is returned
}

func TestRendezvous(t *testing.T) {
	size := 10
	nodes := generateNodes(size, 0, 3)
	if len(nodes) == 0 {
		t.Fatal("generate nodes' length is 0")
	}
	base, follower := nodes[0], nodes[1:]
	base.createNewRing()
	_ = createNetworkGraphFile("grow_ring_create.dot", nodes)
	for i, n := range follower {
		oldPre := base.predecessor
		n.joinRing(base)
		if i == size-1 {
			_ = createNetworkGraphFile(fmt.Sprintf("new_ring_join.dot", i), nodes)
		}
		for i := 0; i < 4; i++ {
			n.fixFigures()
			n.stabilize()
		}
		base.stabilize()
		for i := 0; i < 4; i++ {
			base.fixFigures()
		}
		base.checkPredecessor()
		if i == size-1 {
			_ = createNetworkGraphFile("new_ring_base-pred-stabiliezed.dot", nodes)
		}
		if oldPre != nil {
			oldPre.stabilize()
			for i := 0; i < 4; i++ {
				oldPre.fixFigures()
			}
			oldPre.checkPredecessor()
		}
		if i == size-1 {
			_ = createNetworkGraphFile(fmt.Sprintf("new_ring_old-pred-stabilized.dot", i), nodes)
		}
	}
	for _, n := range nodes {
		n.checkPredecessor()
		for i := 0; i < 4; i++ {
			n.fixFigures()
			n.stabilize()
		}
	}
	_ = createNetworkGraphFile("new_ring_completed.dot", nodes)
	nodes[0].failed = true
	for _, n := range nodes {
		n.checkPredecessor()
		for i := 0; i < 4; i++ {
			n.fixFigures()
			n.stabilize()
		}
	}
	_ = createNetworkGraphFile("new_ring_one_failed.dot", nodes)
}

func createNetworkGraphFile(name string, nodes []*Node) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	fmt.Fprintln(f, networkGraph(nodes))
	f.Close()
	return nil
}
func networkGraph(rings []*Node) string {
	describe := bytes.NewBuffer([]byte{})
	for _, n := range rings {
		// add successors
		for _, s := range n.successors {
			if s == nil {
				continue
			}
			describe.WriteString(
				fmt.Sprintf("    \"id:%d\" -> \"id:%d\" [weight = 100, label = %s, color = deeppink];\n",
					uint64(n.id), uint64(s.id), "succ"),
			)
		}
		// add finger
		for _, f := range n.finger {
			if f == nil {
				continue
			}
			describe.WriteString(
				fmt.Sprintf("    \"id:%d\" -> \"id:%d\" [style = \"dotted\", arrowsize = 0.5, coloer = gray80];\n", uint64(n.id), uint64(f.id)),
			)
		}
		if p := n.predecessor; p != nil {
			describe.WriteString(
				fmt.Sprintf("    \"id:%d\" -> \"id:%d\" [weight = 100, label = %s, color = deepskyblue1];\n",
					uint64(n.id), uint64(p.id), "predecessor"),
			)
		}
	}
	return fmt.Sprintf("digraph network {\n %s}\n", describe.String())
}

func generateTestHash(sup uint64) func(k string) uint64 {
	return func(k string) uint64 {
		u, err := strconv.ParseUint(k, 16, 64)
		if err != nil {
			return 0 % sup
		}
		return u % sup
	}
}
func testHash(k string) uint64 {
	u, err := strconv.ParseUint(k, 16, 64)
	if err != nil {
		return 0
	}
	return u
}

func generateNodes(length, offset, step int) []*Node {
	ring := []*Node{}
	for i := 0; i < length; i++ {
		ring = append(ring, NewNode(fmt.Sprintf("%x", offset+i*step), 4, generateTestHash(uint64(offset+length*step))))
	}
	return ring
}
func setupRingStatically(ring []*Node, sl int) {
	// finger table size is 1
	l := len(ring)
	if sl > l {
		panic("successors len is bigger to length of given ring")
	}
	for k := range ring {
		// register self information
		current := ring[k]
		// register predecessor
		if k == 0 {
			ring[k].predecessor = ring[l-1]
		} else {
			current.predecessor = ring[k-1]
		}
		if k == l-1 {
			ring[k].finger = append(ring[k].finger, ring[0])
		} else {
			current.finger = append(current.finger, ring[k+1])
		}
		// register successors
		for i := k + 1; i < k+sl+1; i++ {
			index := i
			if !(i < l) {
				index = i % l
			}
			current.successors = append(current.successors, ring[index])
		}
	}
}

func assertTestRingStatically(t *testing.T, ring []*Node) {
	t.Run("assert successors' relation", func(t *testing.T) {
		for k := range ring {
			if k+1 < len(ring) {
				if ring[k].successors[0] != ring[k+1] {
					t.Errorf("node's succesor is next node")
				}
			}
		}
		if ring[len(ring)-1].successors[0] != ring[0] {
			t.Errorf("node's succesor is next node")
		}
	})
	t.Run("assert predecessor's relation", func(t *testing.T) {
		for k := range ring {
			if k > 0 {
				if ring[k].predecessor != ring[k-1] {
					t.Errorf("node's predecessor is prev node")
				}
			}
		}
		if ring[0].predecessor != ring[len(ring)-1] {
			t.Errorf("node's predecessor is prev node")
		}
	})
}
