package main

import (
	"televito-parser/addSources/myAutoGe"
)

func main() {
	var myautoGePage int32 = 1
	for {
		myautoGePage = myAutoGe.ParsePage(myautoGePage)
	}
}
