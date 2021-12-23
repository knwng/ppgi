package intersect

type Intersecter interface {
	Run() error
	runClient() error
	runHost() error
}
