package container

type IErrorLogger interface {
	Error(error, string)
}
