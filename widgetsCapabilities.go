package main

type Searcher interface {
	Search(string, bool)
	SearchNext()
	SearchPrev()
	SearchClear()
}

type Mover interface {
	MoveUp()
	MoveDown()
	MoveLeft()
	MoveRight()
}

type Submitter interface {
	Submit()
}
