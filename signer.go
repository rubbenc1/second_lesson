package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, j := range jobs {
		out := make(chan interface{}) // Создаем НОВЫЙ канал для каждой стадии
		wg.Add(1)

		go func(currentJob job, input, output chan interface{}) {
			defer wg.Done()
			defer close(output) // Важно: закрываем выход, чтобы следующая стадия вышла из range
			currentJob(input, output)
		}(j, in, out)
		in = out // Выход текущей стадии становится входом для следующей
	}

	wg.Wait()
}

var md5Lock sync.Mutex // Глобальный или переданный мьютекс

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {
		wg.Add(1)

		go func(data interface{}) {
			defer wg.Done()
			strData := fmt.Sprintf("%v", data)

			// 1. Сначала вычисляем MD5 (строго последовательно)
			md5Lock.Lock()
			md5Data := DataSignerMd5(strData)
			md5Lock.Unlock()

			// 2. Затем запускаем CRC32 (их можно вызывать параллельно)
			crc32Chan := make(chan string)
			crc32Md5Chan := make(chan string)

			go func() { crc32Chan <- DataSignerCrc32(strData) }()
			go func() { crc32Md5Chan <- DataSignerCrc32(md5Data) }()

			// 3. Собираем результат
			res := (<-crc32Chan) + "~" + (<-crc32Md5Chan)
			out <- res
		}(input)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for singleHashResult := range in {
		wg.Add(1)
		go func(result interface{}) {
			defer wg.Done()
			innerWg := &sync.WaitGroup{}
			collection := make([]string, 6)
			for i := 0; i <= 5; i++ {
				innerWg.Add(1)
				concatenated := fmt.Sprintf("%v%v", i, result)
				go func(index int) {
					collection[index] = DataSignerCrc32(concatenated)
					defer innerWg.Done()
				}(i)
			}
			innerWg.Wait()
			out <- strings.Join(collection, "")
		}(singleHashResult)

	}
	wg.Wait()

}

func CombineResults(in, out chan interface{}) {
	var results []string
	for res := range in {
		results = append(results, res.(string))
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}

func main() {
}
