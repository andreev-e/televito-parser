package main

import (
	"./addSources/myAutoGe"
)

func main() {
	var myAutoGePage int32 = 1
	for {
		myAutoGePage = myAutoGe.ParsePage(myAutoGePage)
		return
	}
}
