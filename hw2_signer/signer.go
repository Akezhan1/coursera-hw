package main

import (
	"strconv"
	"sync"
)

// сюда писать код

func main() {

	inputData := []int{0, 1}
	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
	}

	ExecutePipeline(hashSignJobs...)

}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 5)
	out := make(chan interface{}, 5)

	for i, f := range jobs {
		if i%2 == 0 {
			go f(in, out)
		} else {
			go f(out, in)
		}
	}
}

func SingleHash(in, out chan interface{}) {
	var data, md5Hash, crc32md5Hash, crc32Hash, resultHash string
	wg := &sync.WaitGroup{}
	for value := range in {
		data = strconv.Itoa(value.(int))
		md5Hash = DataSignerMd5(data)

		wg.Add(1)
		go func(crc32 *string, input string, wg *sync.WaitGroup) {
			*crc32 = DataSignerCrc32(input)
			defer wg.Done()
		}(&crc32Hash, data, wg)

		wg.Add(1)
		go func(crc32md5 *string, input string, wg *sync.WaitGroup) {
			*crc32md5 = DataSignerCrc32(input)
			defer wg.Done()
		}(&crc32md5Hash, md5Hash, wg)

		wg.Wait()

		resultHash = crc32Hash + "~" + crc32md5Hash
		//fmt.Printf("single hash: %s\n", resultHash)
		out <- resultHash

	}
}

func MultiHash(in, out chan interface{}) {

}

func CombineResults(in, out chan interface{}) {

}
