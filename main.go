package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingBuffer struct {
	arr  []int
	pos  int //текущая позиция буфера
	size int //размер буфера
	mtx  sync.Mutex
}

const bufferSize int = 10
const bufferClearInterval = 10 * time.Second

func main() {
	inputCh := make(chan int)
	done := make(chan bool)
	go read(inputCh, done)

	//фильтр отрицательных чисел
	negativeNumFilterChannel := make(chan int)
	go filterNegativeValues(inputCh, negativeNumFilterChannel, done)

	//фильтр чисел не кратных 3(не пропускать), исключая и 0
	notDevidedFilterChannel := make(chan int)
	go filterNotDevidedValues(negativeNumFilterChannel, notDevidedFilterChannel, done)

	//буферизация данных
	bufferChannel := make(chan int)
	go bufferStage(notDevidedFilterChannel, bufferChannel, done)

	for {
		select {
		case data := <-bufferChannel:
			fmt.Println("Данные ", data)
		case <- done:
			return
		}
	}
}

func read(next chan<- int, done chan bool) {
	var data string
	fmt.Println("Введите значение")
	//считать значение из терминала
	scan := bufio.NewScanner(os.Stdin)

	for scan.Scan() {
		data = scan.Text()
		fmt.Println("введено значение ", data)
		if strings.EqualFold(data, "exit") {
			fmt.Println("Работа завершена")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("необходимо ввести целое число")
			continue
		}
		next <- i
	}
}

func filterNegativeValues(prevChan <-chan int, nextChan chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-prevChan:
			if data > 0 {
				fmt.Println("значение больше 0")
				nextChan<- data
			}
		case <-done:
			return
		}
	}
}

func filterNotDevidedValues(prevChan <-chan int, nextChan chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-prevChan:
			if data%3 == 0 && data != 0 {
				fmt.Println("значение кратно 3")
				nextChan <- data
			}
		case <-done:
			return
		}
	}
}

func bufferStage(prevChan <-chan int, nextChan chan<- int, done <-chan bool) {
	buffer := NewRingBuffer(bufferSize)
	for {
		select {
		case data := <-prevChan:
			buffer.Push(data)
			fmt.Println("добавлено значение в буфер ", data)
		case <-time.After(bufferClearInterval):
			bufferData := buffer.Get()
			if bufferData != nil {
				for _, data := range bufferData {
					nextChan <- data
				}
			}
		case <-done:
			return
		}
	}
}
 
//конструктор кольцевого буфера
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

//добавить элемент в буфер
func (r *RingBuffer) Push(elem int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if r.pos == r.size - 1 {
		for i := 1; i <= r.size - 1; i++ {
			r.arr[i-1] = r.arr[i]
		}
		r.arr[r.pos] = elem
	} else {
		r.pos++
		r.arr[r.pos] = elem
	}
}

//получить элементы из буфера
func (r *RingBuffer) Get() []int {
	if (r.pos <= 0) {
		return nil
	}
	r.mtx.Lock()
	defer r.mtx.Unlock()
	var output []int = r.arr[:r.pos+1]
	r.pos = -1
	return output
}

