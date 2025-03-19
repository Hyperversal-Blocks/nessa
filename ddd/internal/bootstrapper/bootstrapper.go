package bootstrapper

type bootstrapper struct{}

type Bootstrapper interface{}

func NewBootstrapper() Bootstrapper {
	return &bootstrapper{}
}
