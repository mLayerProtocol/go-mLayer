package ring

import (
	"fmt"
	"sync"

	"github.com/mlayerprotocol/go-mlayer/pkg/log"

	"github.com/zeebo/xxh3"
)
var logger = &log.Logger
var GlobalHashRing *HashRing
// Node represents a physical node in the system.
type Node struct {
	ID     string
	PubKey string
	VirtualNodes []*VirtualNode
}

// VirtualNode represents a virtual node assigned to a physical node.
type VirtualNode struct {
	ID          string
	Node        *Node
	PrimaryNode *Node
}

// HashRing represents the consistent hash ring.
type HashRing struct {
	nodeMap   map[uint64]*VirtualNode
	nodes     []*Node
	mutex     sync.RWMutex
	virtuals  int
}



// NewHashRing initializes a hash ring with a predefined list of nodes.
func NewHashRing(physicalNodes []*Node, virtuals int) (*HashRing, error) {
	nodes := make([]*Node, len(physicalNodes))

	copy(nodes, physicalNodes)
	logger.Infof("HASHRINGNODES %v %v", nodes, physicalNodes)
	return (&HashRing{
		nodeMap:  make(map[uint64]*VirtualNode),
		nodes:    nodes,
		virtuals: virtuals,
	}).distributeNodes()
}

// distributeNodes assigns each node's virtual nodes to the next two nodes (n+1, n+2).
func (h *HashRing) distributeNodes() (*HashRing, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	totalNodes := len(h.nodes)
	if totalNodes < 2 {
		logger.Println("Not enough nodes to assign virtual nodes. Need at least 2. got ", totalNodes)
		return h, fmt.Errorf("not enough nodes to assign virtual nodes. need at least 3, got %d", totalNodes)
	}

	for i, node := range h.nodes {
		// Virtual nodes assigned to n+1 and n+2 (circular)
		vn1 := h.nodes[(i+1)%totalNodes] // Next node
		vn2 := h.nodes[(i+2)%totalNodes] // Node after next
		if node.VirtualNodes == nil {
			node.VirtualNodes = []*VirtualNode{}
		}
		
		// Create virtual nodes and hash them
		vn1ID := fmt.Sprintf("%s-VN1", node.ID)
		vn2ID := fmt.Sprintf("%s-VN2", node.ID)

		vn1Hash := h.hash(vn1ID)
		vn2Hash := h.hash(vn2ID)

		h.nodeMap[vn1Hash] = &VirtualNode{ID: vn1ID, Node: vn1, PrimaryNode: node}
		h.nodeMap[vn2Hash] = &VirtualNode{ID: vn2ID, Node: vn2, PrimaryNode: node}
		node.VirtualNodes = append(node.VirtualNodes, h.nodeMap[vn1Hash] , h.nodeMap[vn2Hash])

		logger.Printf("Node %s has virtual nodes mapped to: %s and %s\n", node.ID, vn1.ID, vn2.ID)
	}
	return h, nil
}

// AddNode adds a new node to the hash ring and redistributes events.
func (h *HashRing) AddNode(newNodeID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Create the new node
	newNode := &Node{ID: newNodeID}
	h.nodes = append(h.nodes, newNode)

	// Add virtual nodes for the new node
	for i := 0; i < h.virtuals; i++ {
		virtualID := fmt.Sprintf("%s-VN%d", newNodeID, i)
		virtualHash := h.hash(virtualID)

		// Find the successor node for this virtual node
		successor := h.findSuccessor(virtualHash)
		if successor == nil {
			return fmt.Errorf("no successor found for virtual node %s", virtualID)
		}

		// Assign the virtual node to the new node
		h.nodeMap[virtualHash] = &VirtualNode{ID: virtualID, Node: newNode, PrimaryNode: successor}

		// Redistribute keys from the successor to the new node
		h.redistributeKeys(successor, newNode, virtualHash)
	}

	logger.Printf("Node %s added to the hash ring\n", newNodeID)
	return nil
}

// RemoveNode removes a node from the hash ring and redistributes its events.
func (h *HashRing) RemoveNode(nodeID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Find the node to remove
	var nodeToRemove *Node
	var index int
	for i, node := range h.nodes {
		if node.ID == nodeID {
			nodeToRemove = node
			index = i
			break
		}
	}
	if nodeToRemove == nil {
		return fmt.Errorf("node %s not found in the hash ring", nodeID)
	}

	// Remove the node from the nodes list
	h.nodes = append(h.nodes[:index], h.nodes[index+1:]...)

	// Remove its virtual nodes from the ring
	for i := 0; i < h.virtuals; i++ {
		virtualID := fmt.Sprintf("%s-VN%d", nodeID, i)
		virtualHash := h.hash(virtualID)

		// Find the successor node for this virtual node
		successor := h.findSuccessor(virtualHash)
		if successor == nil {
			return fmt.Errorf("no successor found for virtual node %s", virtualID)
		}

		// Redistribute keys from the removed node to the successor
		h.redistributeKeys(nodeToRemove, successor, virtualHash)

		// Remove the virtual node from the ring
		delete(h.nodeMap, virtualHash)
	}

	logger.Printf("Node %s removed from the hash ring\n", nodeID)
	return nil
}

// findSuccessor finds the node responsible for the given hash.
func (h *HashRing) findSuccessor(hash uint64) *Node {
	var successor *Node
	var minDiff uint64 = ^uint64(0) // Max uint64

	for nodeHash, vn := range h.nodeMap {
		diff := absDiff(hash, nodeHash)
		if diff < minDiff {
			minDiff = diff
			successor = vn.Node
		}
	}

	return successor
}

// redistributeKeys transfers keys from one node to another.
func (h *HashRing) redistributeKeys(from, to *Node, newHash uint64) {
	// Simulate transferring keys
	logger.Printf("Transferring keys from node %s to node %s\n", from.ID, to.ID)

	// In a real implementation, you would:
	// 1. Fetch keys from the `from` node that fall within the `to` node's range.
	// 2. Assign these keys to the `to` node.
}

// GetNode finds the physical node responsible for the given key.
func (h *HashRing) GetNode(key string) *Node {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(h.nodeMap) == 0 {
		logger.Println("No nodes available in the hash ring.")
		return nil
	}

	hash := h.hash(key)
	// Find the nearest virtual node using linear search
	var closestNode *Node
	var minDiff uint64 = ^uint64(0) // Max uint64

	for nodeHash, vn := range h.nodeMap {
		diff := absDiff(hash, nodeHash)
		if diff < minDiff {
			minDiff = diff
			closestNode = vn.Node
		}
	}

	return closestNode
}

// hash generates a 64-bit hash using XXH3.
func (h *HashRing) hash(key string) uint64 {
	return xxh3.HashString(key)
}

// absDiff computes the absolute difference between two uint64 numbers.
func absDiff(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return b - a
}



// package ring

// import (
// 	"crypto/tls"
// 	"fmt"
// 	"sort"
// 	"sync"

// 	"github.com/mlayerprotocol/go-mlayer/entities"
// 	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
// 	"github.com/mlayerprotocol/go-mlayer/pkg/logger"
// 	"github.com/zeebo/xxh3"
// )

// var logger = &logger.Logger


// // Node represents a node in the network
// type Node struct {
// 	ID       string
// 	// Address  string
// 	VNodes   []VirtualNode
// 	Position uint32
// 	// Events   []entities.Event    // Events this node is responsible for
// 	datastore   *ds.Datastore    // Datastore
// 	mu       sync.Mutex // Protects Events
// }

// // VirtualNode represents a virtual node for redundancy
// type VirtualNode struct {
// 	ParentID string
// 	ID string
// 	Position uint32
// 	Events   []entities.Event    // Events this virtual node is backing up
// 	datastore   *ds.Datastore   
// 	mu       sync.Mutex // Protects Events
// }

// // Ring represents the consistent hash ring
// type Ring struct {
// 	nodes     map[uint32]*Node
// 	vNodes    map[uint32]*VirtualNode
// 	positions []uint32
// 	mu        sync.RWMutex
// }

// // NewRing creates a new consistent hash ring
// func NewRing() *Ring {
// 	return &Ring{
// 		nodes:     make(map[uint32]*Node),
// 		vNodes:    make(map[uint32]*VirtualNode),
// 		positions: make([]uint32, 0),
// 	}
// }
// func (r *Ring) BatchAddNodes(nodes []*Node, numVNodesPerNode int) map[uint32][]uint32 {
//     r.mu.Lock()
//     defer r.mu.Unlock()

//     // Map to track which existing positions are affected by which new positions
//     affectedPositions := make(map[uint32][]uint32)
//     newPositions := make([]uint32, 0)

//     // First, calculate all new positions without modifying the ring
//     for _, node := range nodes {
//         // Calculate physical node position
//         position := r.hash(node.ID)
//         newPositions = append(newPositions, position)

//         // Calculate virtual node positions
//         for i := 0; i < numVNodesPerNode; i++ {
//             vNodeID := fmt.Sprintf("%s-vnode-%d", node.ID, i)
//             vNodePos := r.hash(vNodeID)
//             newPositions = append(newPositions, vNodePos)
//         }
//     }

//     // Sort all new positions
//     sort.Slice(newPositions, func(i, j int) bool {
//         return newPositions[i] < newPositions[j]
//     })

//     // Find affected positions for each new position
//     for _, newPos := range newPositions {
//         affected := r.findAffectedPositions(newPos)
//         affectedPositions[newPos] = affected
//     }

//     // Now actually add the nodes and their virtual nodes
//     for _, node := range nodes {
//         position := r.hash(node.ID)
//         node.Position = position
//         r.nodes[position] = node
//         r.positions = append(r.positions, position)

//         for i := 0; i < numVNodesPerNode; i++ {
//             vNodeID := fmt.Sprintf("%s-vnode-%d", node.ID, i)
//             vNodePos := r.hash(vNodeID)
// 			// TODO - pick random nodes from the existing primary nodes
//             vNode := VirtualNode{
//                 ParentID: node.ID,
//                 Position: vNodePos,
//             }
//             r.vNodes[vNodePos] = &vNode
//             r.positions = append(r.positions, vNodePos)
//         }
//     }

//     // Final sort of all positions
//     sort.Slice(r.positions, func(i, j int) bool {
//         return r.positions[i] < r.positions[j]
//     })

//     return affectedPositions
// }

// // AddNode adds a single node to the ring and returns affected positions
// func (r *Ring) AddNode(node *Node, numVNodes int) []uint32 {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	// Calculate new node's position
// 	position := r.hash(node.ID)
	
// 	// Find the position this new node will take over from
// 	affectedPos := r.findAffectedPositions(position)
	
// 	// Add physical node
// 	node.Position = position
// 	r.nodes[position] = node
// 	r.positions = append(r.positions, position)

// 	// Add virtual nodes
// 	vNodePositions := make([]uint32, 0, numVNodes)
// 	for i := 0; i < numVNodes; i++ {
// 		vNodeID := fmt.Sprintf("%s-vnode-%d", node.ID, i)
// 		vNodePos := r.hash(vNodeID)
// 		vNode := VirtualNode{
// 			ParentID: node.ID,
// 			Position: vNodePos,
// 		}
// 		r.vNodes[vNodePos] = &vNode
// 		r.positions = append(r.positions, vNodePos)
// 		vNodePositions = append(vNodePositions, vNodePos)
		
// 		// Find positions affected by virtual nodes
// 		affectedPos = append(affectedPos, r.findAffectedPositions(vNodePos)...)
// 	}

// 	sort.Slice(r.positions, func(i, j int) bool {
// 		return r.positions[i] < r.positions[j]
// 	})

// 	return affectedPos
// }

// // findAffectedPositions finds the position that will be affected by adding a new position
// func (r *Ring) findAffectedPositions(newPos uint32) []uint32 {
// 	if len(r.positions) == 0 {
// 		return nil
// 	}

// 	// Find the next position in the ring
// 	idx := sort.Search(len(r.positions), func(i int) bool {
// 		return r.positions[i] > newPos
// 	})
// 	if idx == len(r.positions) {
// 		idx = 0
// 	}

// 	// Return the position that will be affected
// 	return []uint32{r.positions[idx]}
// }

// // RemoveNode removes a node and returns positions that need redistribution
// func (r *Ring) RemoveNode(nodeID string) []uint32 {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	var affectedPositions []uint32

// 	// Find the node's position
// 	var nodePos uint32
// 	for pos, node := range r.nodes {
// 		if node.ID == nodeID {
// 			nodePos = pos
// 			// Find who takes over this position
// 			affectedPositions = append(affectedPositions, r.findNextPosition(nodePos))
// 			delete(r.nodes, pos)
// 			break
// 		}
// 	}

// 	// Remove virtual nodes and find affected positions
// 	var vNodePositions []uint32
// 	for pos, vNode := range r.vNodes {
// 		if vNode.ParentID == nodeID {
// 			vNodePositions = append(vNodePositions, pos)
// 			// Find who takes over each virtual node position
// 			affectedPositions = append(affectedPositions, r.findNextPosition(pos))
// 			delete(r.vNodes, pos)
// 		}
// 	}

// 	// Update positions slice
// 	newPositions := make([]uint32, 0, len(r.positions)-len(vNodePositions)-1)
// 	for _, pos := range r.positions {
// 		if pos != nodePos && !contains(vNodePositions, pos) {
// 			newPositions = append(newPositions, pos)
// 		}
// 	}
// 	r.positions = newPositions

// 	return affectedPositions
// }

// // findNextPosition finds the next position in the ring that will take over
// func (r *Ring) findNextPosition(pos uint32) uint32 {
// 	idx := sort.Search(len(r.positions), func(i int) bool {
// 		return r.positions[i] > pos
// 	})
// 	if idx == len(r.positions) {
// 		idx = 0
// 	}
// 	return r.positions[idx]
// }

// // NodeChange represents a change in node membership
// type NodeChange struct {
// 	Type     string // "join" or "leave"
// 	Node     *Node
// 	NumVNode int
// }

// // EventRedistribution represents events that need to be redistributed
// type EventRedistribution struct {
// 	Event    *entities.Event
// 	FromNode string
// 	ToNode   string
// }

// // eventDistributor handles the distribution of events
// type eventDistributor struct {
// 	ring              *Ring
// 	membershipChanges chan NodeChange
// 	eventRedist      chan EventRedistribution
// 	nodes            map[string]*Node // Quick lookup by node ID
// 	mu               sync.RWMutex
// }

// // NeweventDistributor creates a new event distributor
// func NewEventDistributor(addr string) (*eventDistributor, error) {
	

// 	ed := &eventDistributor{
// 		ring:              NewRing(),
// 		membershipChanges: make(chan NodeChange, 100),
// 		eventRedist:      make(chan EventRedistribution, 1000),
// 		nodes:            make(map[string]*Node),
// 	}

// 	// Start membership change handler
// 	go ed.handleMembershipChanges()
// 	// Start event redistribution handler
// 	go ed.handleEventRedistribution()

// 	return ed, nil
// }

// // handleMembershipChanges processes node joins and leaves
// func (ed *eventDistributor) handleMembershipChanges() {
// 	for change := range ed.membershipChanges {
// 		switch change.Type {
// 		case "join":
// 			ed.handleNodeJoin(change.Node, change.NumVNode)
// 		case "leave":
// 			ed.handleNodeLeave(change.Node)
// 		}
// 	}
// }
// // hash generates a consistent hash for a given string
// func (r *Ring) hash(key string) uint32 {
// 	h := xxh3.HashString(key)
// 	return uint32(h) ^ uint32(h>>32)
// }


// // GetNode returns the node responsible for handling the event
// func (r *Ring) GetNode(eventID string) (*Node, []*VirtualNode) {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()

// 	if len(r.positions) == 0 {
// 		return &Node{}, nil
// 	}

// 	// Find the position for this event
// 	hash := r.hash(eventID)
// 	idx := sort.Search(len(r.positions), func(i int) bool {
// 		return r.positions[i] >= hash
// 	})

// 	if idx == len(r.positions) {
// 		idx = 0
// 	}

// 	position := r.positions[idx]
// 	primaryNode := r.nodes[position]

// 	// Get the next two virtual nodes for redundancy
// 	backupVNodes := make([]*VirtualNode, 0, 2)
// 	seen := make(map[string]bool)
// 	seen[primaryNode.ID] = true

// 	for i := 1; len(backupVNodes) < 2 && i < len(r.positions); i++ {
// 		nextPos := r.positions[(idx+i)%len(r.positions)]
// 		if vNode, ok := r.vNodes[nextPos]; ok {
// 			if !seen[vNode.ParentID] {
// 				backupVNodes = append(backupVNodes, vNode)
// 				seen[vNode.ParentID] = true
// 			}
// 		}
// 	}

// 	return primaryNode, backupVNodes
// }


// // handleEventRedistribution processes event redistribution
// func (ed *eventDistributor) handleEventRedistribution() {
// 	for redist := range ed.eventRedist {
// 		if err := ed.redistributeEvent(redist); err != nil {
// 			logger.Printf("Failed to redistribute event %s: %v", redist.Event.ID, err)
// 		}
// 	}
// }

// // redistributeEvent sends an event to its new owner
// func (ed *eventDistributor) redistributeEvent(redist EventRedistribution) error {
// 	ed.mu.RLock()
// 	toNode, exists := ed.nodes[redist.ToNode]
// 	ed.mu.RUnlock()

// 	if !exists {
// 		return fmt.Errorf("destination node %s not found", redist.ToNode)
// 	}

// 	// Send event to new primary node
// 	// if err := ed.sendToNode(toNode, redist.Event); err != nil {
// 	// 	return fmt.Errorf("failed to send event to new node: %v", err)
// 	// }

// 	// Update node's event list
// 	toNode.mu.Lock()
// 	// toNode.Events = append(toNode.Events, redist.Event)
// 	toNode.mu.Unlock()

// 	return nil
// }
// // handleNodeJoin processes a new node joining the network
// // BatchNodeChange represents multiple nodes joining or leaving
// type BatchNodeChange struct {
//     Type      string // "join" or "leave"
//     Nodes     []Node
//     NumVNodes int
// }

// func (ed *eventDistributor) handleBatchNodeJoin(nodes []*Node, numVNodes int) {
//     ed.mu.Lock()
//     defer ed.mu.Unlock()

//     // Add all nodes and get affected positions
//     affectedPositionsMap := ed.ring.BatchAddNodes(nodes, numVNodes)

//     // Create a map for quick node lookup
//     newNodeMap := make(map[string]bool)
//     for _, node := range nodes {
//         ed.nodes[node.ID] = node
//         newNodeMap[node.ID] = true
//     }

//     // Track all events that need redistribution
//     redistributions := make(map[string][]EventRedistribution)

//     // Process all affected positions
//     for _, affectedPositions := range affectedPositionsMap {
//         for _, affectedPos := range affectedPositions {
//             if affectedNode, exists := ed.nodes[ed.ring.nodes[affectedPos].ID]; exists {
//                 affectedNode.mu.Lock()
                
//                 // Check each event on the affected node
//                 // var remainingEvents []*entities.Event
//                 // for _, event := range affectedNode.Events {
//                 //     newPrimaryNode, _ := ed.ring.GetNode(event.ID)
                    
//                 //     // If event should move to a new node
//                 //     if newNodeMap[newPrimaryNode.ID] {
//                 //         redistributions[newPrimaryNode.ID] = append(
//                 //             redistributions[newPrimaryNode.ID],
//                 //             EventRedistribution{
//                 //                 Event:    event,
//                 //                 FromNode: affectedNode.ID,
//                 //                 ToNode:   newPrimaryNode.ID,
//                 //             },
//                 //         )
//                 //     } else {
//                 //         remainingEvents = append(remainingEvents, event)
//                 //     }
//                 // }
                
//                 // (*affectedNode).Events = remainingEvents
//                //  affectedNode.mu.Unlock()
//             }
//         }
//     }

//     // Perform redistributions in parallel
//     var wg sync.WaitGroup
//     for nodeID, redists := range redistributions {
//         wg.Add(1)
//         go func(nID string, redists []EventRedistribution) {
//             defer wg.Done()
//             for _, redist := range redists {
//                 ed.eventRedist <- redist
//             }
//         }(nodeID, redists)
//     }
//     wg.Wait()
// }

// func (ed *eventDistributor) handleNodeJoin(node *Node, numVNodes int) {
// 	ed.mu.Lock()
// 	defer ed.mu.Unlock()

// 	// Add node and get affected positions
// 	affectedPositions := ed.ring.AddNode(node, numVNodes)
// 	ed.nodes[node.ID] = node
// 	logger.Infof("AddedNode: %s", node.ID)
// 	// Only redistribute events from affected positions
// 	for _, pos := range affectedPositions {
// 		logger.Infof("RINGPOSITION: %d:%v", pos, ed.ring.nodes[pos])
		

// 		if affectedNode, exists := ed.nodes[ed.ring.nodes[pos].ID]; exists {
// 			affectedNode.mu.Lock()
// 			//var eventsToRedistribute []*entities.Event
			
// 			// Only redistribute events that now belong to the new node
// 			// for _, event := range affectedNode.Events {
// 			// 	newPrimaryNode, _ := ed.ring.GetNode(event.ID)
// 			// 	if newPrimaryNode.ID == node.ID {
// 			// 		eventsToRedistribute = append(eventsToRedistribute, event)
// 			// 		// Remove event from old node
// 			// 		affectedNode.Events = removeEvent(affectedNode.Events, event)
// 			// 	}
// 			// }
// 			defer affectedNode.mu.Unlock()

// 			// Redistribute affected events
// 			// for _, event := range eventsToRedistribute {
// 			// 	ed.eventRedist <- EventRedistribution{
// 			// 		Event:    event,
// 			// 		FromNode: affectedNode.ID,
// 			// 		ToNode:   node.ID,
// 			// 	}
// 			// }
// 		}
// 	}
// }

// // handleNodeLeave processes a node leaving the network
// func (ed *eventDistributor) handleNodeLeave(node *Node) {
// 	ed.mu.Lock()
// 	defer ed.mu.Unlock()

// 	// Get affected positions and their events
// 	// affectedPositions := ed.ring.RemoveNode(node.ID)
	
// 	// if leavingNode, exists := ed.nodes[node.ID]; exists {
// 	// 	leavingNode.mu.Lock()
// 	// 	eventsToRedistribute := leavingNode.Events
// 	// 	leavingNode.mu.Unlock()

// 	// 	// Redistribute events to new owners
// 	// 	for _, event := range eventsToRedistribute {
// 	// 		newPrimaryNode, _ := ed.ring.GetNode(event.ID)
// 	// 		ed.eventRedist <- EventRedistribution{
// 	// 			Event:    event,
// 	// 			FromNode: node.ID,
// 	// 			ToNode:   newPrimaryNode.ID,
// 	// 		}
// 	// 	}
// 	// }

// 	delete(ed.nodes, node.ID)
// }

// // Helper function to remove an event from a slice
// func removeEvent(events []*entities.Event, event *entities.Event) []*entities.Event {
// 	for i, e := range events {
// 		if e.ID == event.ID {
// 			return append(events[:i], events[i+1:]...)
// 		}
// 	}
// 	return events
// }

// // HandleEvent processes an incoming event
// func (ed *eventDistributor) HandleEvent(event *entities.Event, sendToNodes func(primaryNode *Node, backupVNodes []*Node) error) error {
// 	primaryNode, backupVNodesList := ed.ring.GetNode(event.ID)
	
// 	// Send to primary node
// 	// if err := ed.sendToNode(primaryNode, *event); err != nil {
// 	// 	logger.Printf("Failed to send to primary node %s: %v", primaryNode.ID, err)
// 	// 	// Try backup nodes
// 	// }
// 	backupVNodes := []*Node{}
// 	for _, vNode := range backupVNodesList {
// 		if node, ok := ed.ring.nodes[vNode.Position]; ok {
// 			backupVNodes = append(backupVNodes, node)
// 		}
// 	}
// 	return sendToNodes(primaryNode, backupVNodes)
// 	// Send to backup virtual nodes
// 	// for _, vNode := range backupVNodes {
// 	// 	if node, ok := ed.ring.nodes[vNode.Position]; ok {
// 	// 		if err := ed.sendToNode(node, event); err != nil {
// 	// 			logger.Printf("Failed to send to backup node %s: %v", node.ID, err)
// 	// 		}
// 	// 	}
// 	// }

// 	// return nil
// }

// // JoinNetwork adds a node to the network
// func (ed *eventDistributor) JoinNetwork(node *Node, numVNodes int) error {
// 	ed.membershipChanges <- NodeChange{
// 		Type:     "join",
// 		Node:     node,
// 		NumVNode: numVNodes,
// 	}
// 	return nil
// }

// // LeaveNetwork removes a node from the network
// func (ed *eventDistributor) LeaveNetwork(node *Node) error {
// 	ed.membershipChanges <- NodeChange{
// 		Type: "leave",
// 		Node: node,
// 	}
// 	return nil
// }

// // func (ed *eventDistributor) sendToNode(node *Node, event entities.Event) error {
// // 	session, err := quic.DialAddr(context.Background(), node.Address, generateTLSConfig(), nil)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to dial node %s: %v", node.ID, err)
// // 	}
// // 	defer session.CloseWithError(0, "")

// // 	stream, err := session.OpenStreamSync()
// // 	if err != nil {
// // 		return fmt.Errorf("failed to open stream: %v", err)
// // 	}
// // 	defer stream.Close()

// // 	// Send event data
// // 	_, err = stream.Write(event.Payload)
// // 	if err != nil {
// // 		return fmt.Errorf("failed to write to stream: %v", err)
// // 	}

// // 	return nil
// // }

// func generateTLSConfig() *tls.Config {
// 	// Implementation omitted for brevity
// 	// You should implement proper TLS configuration here
// 	return &tls.Config{
// 		InsecureSkipVerify: true, // Note: Don't use this in production
// 	}
// }
// // Helper function to check if a slice contains a value
// func contains(slice []uint32, val uint32) bool {
// 	for _, item := range slice {
// 		if item == val {
// 			return true
// 		}
// 	}
// 	return false
// }

 
// var EventDistribtor = &eventDistributor{}

// // func main() {
// // 	// Initialize the distributor
// // 	distributor, err := NeweventDistributor(":9000")
// // 	if err != nil {
// // 		logger.Fatalf("Failed to create distributor: %v", err)
// // 	}

// // 	// Add initial nodes
// // 	node1 := Node{ID: "node1", Address: ":9001"}
// // 	node2 := Node{ID: "node2", Address: ":9002"}
	
// // 	// Join network
// // 	distributor.JoinNetwork(&node1, 2)
// // 	distributor.JoinNetwork(&node2, 2)

// // 	// Example event handling
// // 	event := Event{
// // 		ID:        "event1",
// // 		Payload:   []byte("test event"),
// // 		Timestamp: time.Now(),
// // 	}

// // 	if err := distributor.HandleEvent(event); err != nil {
// // 		logger.Printf("Failed to handle event: %v", err)
// // 	}

// // 	// Simulate node leaving
// // 	time.Sleep(2 * time.Second)
// // 	distributor.LeaveNetwork(&node1)

// // 	// Add new node
// // 	node3 := Node{ID: "node3", Address: ":9003"}
// // 	distributor.JoinNetwork(&node3, 2)
// // }

// // Rest of the code (HandleEvent, sendToNode, etc.) remains the same as in the previous version