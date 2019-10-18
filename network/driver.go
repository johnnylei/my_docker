package network

type Driver interface {
	Create(string, string) error
	Delete(string) error
	GetNetwork() *Network
}
