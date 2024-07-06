package entity

type PageInfo struct {
	HasNextPage bool
	EndCursor   *string
}

type TaskEdge struct {
	Cursor string
	Node   Task
}

type Tasks struct {
	Edges    []TaskEdge
	PageInfo PageInfo
}
