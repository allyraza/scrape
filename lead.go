package main

import (
	"fmt"
)

type Lead struct {
    Name string
    Email string
    Phone string
    Url string
}

func (l Lead) Print() {
    fmt.Printf("%s,%s,%s,%s\n", l.Name, l.Url, l.Phone, l.Email)
}