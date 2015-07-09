package main
//import "fmt"

func main() {

	var contextStore = NewContext()

	contextStore.UpdateContext()

	GetInstanceMapping(contextStore)

	Collector()

}