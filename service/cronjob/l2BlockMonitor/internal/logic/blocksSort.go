package logic

type blockInfosByBlockHeight []*Block

func (p blockInfosByBlockHeight) Len() int {
	return len(p)
}

func (p blockInfosByBlockHeight) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p blockInfosByBlockHeight) Less(i, j int) bool {
	return p[i].BlockHeight < p[j].BlockHeight
}
