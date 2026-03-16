package main

import (
	"fmt"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	defer close(out)
	for input := range in {
		data := fmt.Sprintf("%v", input)

		crc32Chan := make(chan string)
		crc32Md5Chan := make(chan string)

		go func(){
			crc32Chan <- DataSignerCrc32(data)
		}()
		go func(){
			md5Data := DataSignerMd5(data)
			crc32Md5Chan <- DataSignerCrc32(md5Data)
		}()

		out <- (<-crc32Chan) + "~" + (<-crc32Md5Chan)
	}
}

func MultiHash(in, out chan interface{}){
	var wg sync.WaitGroup
	for singleHashResult := range in {
		wg.Add(1)
		go func(result interface{}){
			defer wg.Done()
			var innerWg sync.WaitGroup
			collection := make([]string, 6)
			for i:=0; i<=5; i++{
				innerWg.Add(1)
				concatenated := fmt.Sprintf("%v%v", i, result)
				go func (index int) {
					collection[index] = DataSignerCrc32(concatenated)
					defer innerWg.Done()
				}(i)
			}
			innerWg.Wait()
			out <- strings.Join(collection, "")
		}(singleHashResult)

	}

	go func(){
		wg.Wait()
		close(out)
	}()

}

func main (){
	fmt.Println("Testing SingleHash")

	in := make(chan interface{})
	out := make(chan interface{})
	out1 := make(chan interface{})
	go SingleHash(in, out)
	go MultiHash(out, out1)
	
	go func(){
		in <-0
		in <-1
		in <-2
		close(in)
	}()
	
	fmt.Println("Results:")
    var results []string
    for input := range out1 {
        results = append(results, fmt.Sprintf("%v", input))
    }
    
    fmt.Println(strings.Join(results, "_"))
}

