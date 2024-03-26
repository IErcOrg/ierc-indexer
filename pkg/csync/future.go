package csync

type Future[Q any, R any] struct {
	request  Q
	response chan *R
	err      chan error
}

func NewFuture[Q any, R any](q Q) *Future[Q, R] {
	return &Future[Q, R]{
		request:  q,
		response: make(chan *R, 1),
		err:      make(chan error),
	}
}

func (f *Future[Q, R]) Send(r *R) {
	f.response <- r
	close(f.response)
	close(f.err)
}

func (f *Future[Q, R]) Err(err error) {
	f.err <- err
	close(f.response)
	close(f.err)
}

func (f *Future[Q, R]) Wait() (*R, error) {
	select {
	case resp := <-f.response:
		return resp, nil
	case err := <-f.err:
		return nil, err
	}
}
