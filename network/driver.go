package network

type Driver interface {
	Name() string
	Create(subnet string, name string)(*Network, error)
	Delete(*Network)
}
