package main

import (
	"fmt"
	"time"

	"github.com/YoungGoofy/cache/internal/cache"
)


func main() {
	cache := cache.New(5 * time.Second, 10 * time.Second, 3)
	cache.Add(1, "hello")
	cache.Add(2, "world")
	cache.Add(3, "how")
	cache.Add(4, "are")
	cache.Add(5, "you")
	cache.AddWithTTL(6, "ok", 0)
	fmt.Println(cache.Get(3))
	// cache.Remove(5)
	fmt.Println(cache.Get(4))
	fmt.Println(cache.GetAll())
	time.Sleep(11 * time.Second)
	fmt.Println(cache.GetAll())
}