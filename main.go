package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/DainerAcosta/cache_opt/cache_opt"
)

func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

func GetFibonacci(n int) (int, error) {
	fmt.Println("Se ejecuto GetFibonacci")
	return Fibonacci(n), nil
}

func main() {
	timeInit := time.Now()
	cache := cache_opt.NewCache(GetFibonacci)
	fibo := []int{44, 42, 40, 41, 42, 38, 40, 40, 41, 43, 43, 43, 44}
	var wg sync.WaitGroup
	for _, n := range fibo {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			start := time.Now()
			value, err := cache.Get(index)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("%d, %s, %d\n", index, time.Since(start), value)
		}(n)
	}
	wg.Wait()
	fmt.Printf("tiempo final %s\n", time.Since(timeInit))
}
