package api

type api struct{}

type Api interface {
}

func New() Api {
	return &api{}
}
