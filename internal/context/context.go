package context

type DnfContext struct {
}

var dnfContext DnfContext

func Init() {

}

func GetSelf() *DnfContext {
	return &dnfContext
}
