package main

import (
	"fmt"
	"github.com/alzabo/mot/pkg"
)

func main() {
	c := mot.NewClient("http://localhost:8080", "admin", "seLrjRkH4")
	fmt.Println(c.TorrentList())

}
