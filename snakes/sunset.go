package snakes

import (
	"container/heap"
	"context"
	"math"

	"github.com/Xe/bsnk/api"
	"within.website/ln"
)

// Sunset is a snake AI based off of the rantings of Ahroo in Discord DM.
type Sunset struct{}

func (Sunset) Ping() (*api.PingResponse, error) {
	return &api.PingResponse{
		APIVersion: "1",
		Color:      "#FFCA54",
		HeadType:   "sand-worm",
		TailType:   "round-bum",
	}, nil
}

// Start kicks off a game.
func (Sunset) Start(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}

// Move selects a target and goes for it.
func (Sunset) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	target := selectGreedy(decoded)

	me := decoded.You.Body
	NodePool := []sunsetNode{
		sunsetNode{
			Node:      me[0],
			Cost:      0,
			TotalCost: 1 + int(sunsetHeuristic(me[0], target)),
			Previous:  -1,
		},
	}
	Queue := make(sunsetNodeQueue, 1)

	Queue[0] = &NodePool[0]
	heap.Init(&Queue)

	BestNode := &NodePool[0]

	for len(Queue) > 0 {
		currNode := heap.Pop(&Queue).(*sunsetNode)

		if sunsetIsGoal(currNode.Node, decoded.Board) {
			BestNode = currNode

			break
		}

		Neighs := sunsetGetNeighbors(currNode.Node, decoded.Board)

		_, currInd := sunsetFindNode(NodePool, currNode.Node)

		for _, currNeigh := range Neighs {
			if currNode.Previous != -1 {
				if currNeigh.Eq(NodePool[currNode.Previous].Node) {
					continue
				}
			}

			new := false

			neighNode, _ := sunsetFindNode(NodePool, currNeigh)

			if neighNode == nil {
				new = true

				neighNode = &sunsetNode{
					Node:      currNeigh,
					Cost:      0,
					TotalCost: 0,
					Previous:  currInd,
				}

				NodePool = append(NodePool, *neighNode)
			}

			heuristic := float64(0)
			currCost := sunsetCost(currNode.Node, currNeigh)
			cost := float64(currNode.Cost) + currCost

			if !sunsetIsGoal(currNeigh, decoded.Board) {
				heuristic = sunsetHeuristic(currNode.Node, target)
			}

			total := cost + heuristic

			if !new {
				if total >= float64(neighNode.TotalCost) {
					continue
				}

				neighNode.Previous = currInd
			}

			neighNode.Cost = int(cost)
			neighNode.TotalCost = int(total)

			fixInd := int(0)

			if new {
				Queue.Push(neighNode)
				fixInd = len(Queue) - 1
			} else {
				for _, currNode := range Queue {
					if currNode.Node.Eq(neighNode.Node) {
						break
					}
					fixInd++
				}
			}

			heap.Fix(&Queue, fixInd)

			if BestNode.TotalCost > int(heuristic) {
				BestNode = neighNode
			}
		}
	}

	trueTargetNode := BestNode

	for trueTargetNode.Previous != -1 {
		trueTargetNode = &NodePool[trueTargetNode.Previous]
	}

	var pickDir string

	ctx = ln.WithF(ctx, logCoords("target", target))
	ctx = ln.WithF(ctx, logCoords("trueTarget", trueTargetNode.Node))
	ctx = ln.WithF(ctx, logCoords("bestNode", BestNode.Node))
	ctx = ln.WithF(ctx, logCoords("my_head", me[0]))

	diff := api.Coord{trueTargetNode.Node.X - me[0].X, trueTargetNode.Node.Y - me[0].Y}

	if math.Abs(float64(diff.X)) > math.Abs(float64(diff.Y)) {
		if diff.X > 0 {
			pickDir = "right"
		} else {
			pickDir = "left"
		}
	} else {
		if diff.Y > 0 {
			pickDir = "down"
		} else {
			pickDir = "up"
		}
	}

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (Sunset) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}

type sunsetNode struct {
	Node      api.Coord
	Cost      int
	TotalCost int
	Previous  int
}

type sunsetNodeQueue []*sunsetNode

func (queue sunsetNodeQueue) Len() int      { return len(queue) }
func (queue sunsetNodeQueue) Swap(i, j int) { queue[i], queue[j] = queue[j], queue[i] }

func (queue sunsetNodeQueue) Less(a, b int) bool {
	return queue[a].TotalCost < queue[b].TotalCost
}

func (queue *sunsetNodeQueue) Push(x interface{}) {
	node := x.(*sunsetNode)
	*queue = append(*queue, node)
}

func (queue *sunsetNodeQueue) Pop() interface{} {
	old := *queue
	n := len(old)
	item := old[n-1]
	*queue = old[0 : n-1]
	return item
}

func sunsetCost(from, to api.Coord) float64 {
	return 1
}

func sunsetHeuristic(from, to api.Coord) float64 {
	l := api.Line{A: from, B: to}
	return l.Distance()
}

func sunsetGetNeighbors(focus api.Coord, board api.Board) []api.Coord {
	var result []api.Coord

	offsets := []api.Coord{
		api.Coord{1, 0},
		api.Coord{0, 1},
		api.Coord{-1, 0},
		api.Coord{0, -1},
	}

	for _, currOffs := range offsets {
		newCoord := api.Coord{focus.X + currOffs.X, focus.Y + currOffs.Y}
		if board.Inside(newCoord) {
			safe := true

			for _, currSnek := range board.Snakes {
				for _, currBod := range currSnek.Body {
					if newCoord.Eq(currBod) {
						safe = false
						break
					}
				}

				if !safe {
					break
				}
			}

			if safe {
				result = append(result, newCoord)
			}
		}
	}

	return result
}

func sunsetIsGoal(focus api.Coord, board api.Board) bool {
	for _, currFood := range board.Food {
		if currFood.Eq(focus) {
			return true
		}
	}

	return false
}

func sunsetFindNode(pool []sunsetNode, focus api.Coord) (*sunsetNode, int) {
	var ind int
	for _, currNode := range pool {
		if currNode.Node == focus {
			return &currNode, ind
		}

		ind++
	}

	return nil, -1
}
