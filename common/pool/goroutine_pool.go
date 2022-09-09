package pool

type Pool struct {
	size       int
	taskChan   chan func()
	cancelChan chan struct{}
}

func NewPool(size int) *Pool {
	return &Pool{
		size:     size,
		taskChan: make(chan func()),
	}
}

func (p *Pool) Start() {
	p.cancelChan = make(chan struct{})
	for i := 0; i < p.size; i++ {
		go func() {
			for {
				select {
				case task := <-p.taskChan:
					if task != nil {
						task()
					}
				case <-p.cancelChan:
					return
				}
			}
		}()
	}
}

func (p *Pool) Submit(task func()) {
	p.taskChan <- task
}

func (p *Pool) Stop() {
	close(p.cancelChan)
}
