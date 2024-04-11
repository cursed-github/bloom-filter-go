package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
)


type BloomFilter struct{
	wordCount int
	wordList []string
	bitArray []bool
	bitarrayLength int
	hashFunctionNum int
	hashFunctionArray []func([]byte) uint64
	filepath string
}

func logger(largs ...interface{}) {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	log.SetOutput(file)
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			log.Println("STRING:", v)
		case error:
			log.Println("ERROR:", v)
			// You can add more specific handling here if needed.
		case panic:
			if v {
				log.Println("ACTION: Panic initiated")
				panic("Panic as requested")
			} else {
				log.Println("Boolean false received, no action taken")
			}
		default:
			log.Println("UNKNOWN TYPE:", v)
		}
	}
}


func (bloomfilter *BloomFilter) hashFunctions() {
	var hashFunctions []func([]byte) uint64

	for i:=0;i<bloomfilter.hashFunctionNum;i++ {
		seed:= uint64(i)
		hashFunctions = append(hashFunctions, func(input []byte) uint64 {
		seedBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBytes, seed)
		hashData := append(seedBytes, input...)
		
		// Compute SHA-256 hash and then convert the first 8 bytes to uint64
		sum := sha256.Sum256(hashData)
		return binary.LittleEndian.Uint64(sum[:8]) % uint64(bloomfilter.bitarrayLength)
		})
	}
	bloomfilter.hashFunctionArray = hashFunctions
}

func (bloomfilter *BloomFilter) extractWords() {
	file , err := os.Open(bloomfilter.filepath)

	if err != nil {
		fmt.Println("error opening file", err)
	}

	logFile, errLogfile := os.Create("wordcount.log")
    if errLogfile != nil {
        fmt.Println("error opening log file", errLogfile)
        return // Exit if log file cannot be opened
    }
    defer logFile.Close() 

	defer file.Close()

	reader := bufio.NewReader(file)
	wordCount := 0
	wordList := make([]string,0)


	for {
		line, err := reader.ReadString('\n')
   	 	words := strings.Fields(line)
		wordList = append(wordList, words[0])
    	wordCount += len(words)

    if err != nil {
        if err == io.EOF {
            break // End of file reached, exit the loop
        }
        fmt.Fprintf(logFile,"Error reading file:", err)
        return// Exit the program on error
    }
	fmt.Fprintf(logFile,"here the variables from inside loop %v %v \n", line, words)
	}

	fmt.Fprintf(logFile,"The file contains %d words\n", wordCount)
	bloomfilter.wordCount = wordCount
	bloomfilter.wordList = wordList
	
}

func (bloomfilter *BloomFilter) calulateParams() {
	m:= CalculateBitArrayLength(bloomfilter.wordCount,0.01)
	n:= CalculateHashFunctions(bloomfilter.wordCount,m)

	bloomfilter.bitarrayLength = m
	bloomfilter.hashFunctionNum = n
} 

func CalculateBitArrayLength(n int, p float64) int {
    m := -float64(n) * math.Log(p) / (math.Log(2) * math.Log(2))
    return int(math.Ceil(m)) // Round up to ensure the bit array is large enough
}

// CalculateHashFunctions calculates the optimal number of hash functions
func CalculateHashFunctions(n int, m int) int {
    k := float64(m) / float64(n) * math.Log(2)
    return int(math.Ceil(k)) // Round up to ensure enough hash functions
}

func (bloomfilter *BloomFilter) writeArrayToFile () {
	
	byteArrayLength := len(bloomfilter.bitArray)/8
	if len(bloomfilter.bitArray)%8 != 0 {
		byteArrayLength++
	}
	byteArray := make([]byte,byteArrayLength)

	for i,b := range bloomfilter.bitArray {
		if b {
			byteIndex := i/8
			byteArray[byteIndex] |= 1 << (i%8)
		}
	}

	err := os.WriteFile("bloomFilter.bin", byteArray, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
}

func (bloomfilter *BloomFilter) setBits () {
	bloomfilter.bitArray = make([]bool,bloomfilter.bitarrayLength)
	for i:=0;i<bloomfilter.wordCount;i++{
		for j:=0;j<bloomfilter.hashFunctionNum;j++{
			index:= bloomfilter.hashFunctionArray[j]([]byte(bloomfilter.wordList[i]))
			bloomfilter.bitArray[index] = true
		}
	}
}

func (bloomfilter *BloomFilter) wordExist(word string) bool {
	for j:=0;j<bloomfilter.hashFunctionNum;j++{
		index:= bloomfilter.hashFunctionArray[j]([]byte(word))
		if !bloomfilter.bitArray[index] {
			return false
		}
		
	}
	return true
}


func main() {
	bloomfilter := BloomFilter{
		filepath: "words.txt",
	}
	fmt.Println("bloom filter is", bloomfilter)

	bloomfilter.extractWords()
	fmt.Println("wordList value", bloomfilter.wordList[1000])
    bloomfilter.calulateParams()
	fmt.Println("bitarraylength and hashfunctions", bloomfilter.bitarrayLength,bloomfilter.hashFunctionNum);
	bloomfilter.hashFunctions();
	fmt.Println("hashfunction array ", bloomfilter.hashFunctionArray[0]([]byte("smita")))
	
	//fmt.Println("bit array value at index", bloomfilter.bitArray[bloomfilter.hashFunctionArray[0]([]byte("smita"))])

	bloomfilter.setBits()
	bloomfilter.writeArrayToFile()
	logger("sfhgljds")
	//logger(string(bloomfilter.wordExist("accurately")))
	fmt.Println("word exist or not",bloomfilter.wordExist("accurately"))
	
}