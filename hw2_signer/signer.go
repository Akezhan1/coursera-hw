package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

func main() {}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	wg := &sync.WaitGroup{}
	worker := func(job func(in, out chan interface{}), in, out chan interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		defer close(out)
		job(in, out)
	}

	for _, f := range jobs {
		wg.Add(1)
		go worker(f, in, out, wg)
		in = out
		out = make(chan interface{}, MaxInputDataLen)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mux := &sync.Mutex{}
	worker := func(f func(data string) string, data string, result chan string, mux *sync.Mutex) {
		if cap(result) == 0 {
			mux.Lock()
			result <- f(data)
			mux.Unlock()
		} else {
			result <- f(data)
		}

	}
	for value := range in {
		data := strconv.Itoa(value.(int))
		wg.Add(1)
		go func(data string, wg *sync.WaitGroup, mux *sync.Mutex) {
			defer wg.Done()

			md5 := make(chan string)
			crc32 := make(chan string, 1)
			crc32md5 := make(chan string, 1)

			go worker(DataSignerMd5, data, md5, mux)

			go worker(DataSignerCrc32, data, crc32, mux)

			go worker(DataSignerCrc32, <-md5, crc32md5, mux)

			result := <-crc32 + "~" + <-crc32md5
			out <- result
		}(data, wg, mux)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	//worker := func(f func(data string) string, data *string)

	for value := range in {
		wg.Add(1)
		data := value.(string)

		go func(data string, wg *sync.WaitGroup) {
			defer wg.Done()

			var crc32data [6]chan string
			for th := 0; th <= 5; th++ {
				crc32data[th] = make(chan string)

				go func(count int, data string, result chan string) {
					result <- DataSignerCrc32(strconv.Itoa(count) + data)
				}(th, data, crc32data[th])
			}

			result := ""
			for th := 0; th <= 5; th++ {
				result += <-crc32data[th]
			}
			out <- result
		}(data, wg)
	}
	wg.Wait()
}
func CombineResults(in, out chan interface{}) {
	var hashes []string

	for value := range in {
		hashes = append(hashes, value.(string))
	}

	sort.Strings(hashes)
	out <- strings.Join(hashes, "_")
}
