package main

import (
    "fmt"
    "github.com/tlopo-go/cookie-exporter/cookies"
)

func main() {
    v,_ := cookies.GetNetscape()
    fmt.Printf("%#s\n", v)
}
