package types

// Repository defines the behavior of a proxy specs repository
type Repository interface {
	FindAll() ([]*Backend, error)
	FindByName(name string) (*Backend, error)
	FindByListenPath(path string) (*Backend, error)
	Exists(def *Backend) (bool, error)
	Add(app *Backend) error
	Remove(name string) error
	FindValidAPIHealthChecks() ([]*Backend, error)
}
