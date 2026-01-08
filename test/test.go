package main

import "sync"

type A struct {
	sync.Mutex
	a struct {
		test    int
		sousous struct {
			sousousous struct {
				salut string
			}
		}
	}
}

type B struct {
	Mutex sync.Mutex
}

func F() {
	var a A
	a.Lock()
}

func main() {

}
