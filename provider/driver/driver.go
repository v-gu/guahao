package driver

type Driver interface {
	Login() error
	Book() error
}
