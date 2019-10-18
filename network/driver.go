package network

type Driver interface {
	Name() string
	Create(string, string) (*Network, error)
	Delete(string) error
}
