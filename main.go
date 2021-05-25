package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/jszwec/csvutil"
	"io"
	"log"
	"os"
)

func main() {
	dictionaryChannel := make(chan []Translation)
	wordsChannel := make(chan []string)

	go ReadWords(wordsChannel)
	go ReadDictionary(dictionaryChannel)

	fmt.Printf("%+v \n", <-dictionaryChannel)
	fmt.Printf("%+v \n", <-wordsChannel)
}

func ReadWords(c chan<- []string) {
	words := make([]string, 1000)
	file, err := os.Open("find_words.txt")
	if err != nil {
		log.Fatal("Couldn't open find words file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		words[i] = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading find words file")
	}

	c <- words
}

func ReadDictionary(c chan []Translation) {
	dictionary := make([]Translation, 1000)
	file, err := os.Open("french_dictionary.csv")
	if err != nil {
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := csvutil.Header(Translation{}, "csv")
	if err != nil {
		log.Fatal("Couldn't parse the headers of CSV file")
	}

	decoder, err := csvutil.NewDecoder(reader, header...)
	if err != nil {
		log.Fatal("Couldn't decode CSV file")
	}

	i := 0
	for {
		var translation Translation
		if err := decoder.Decode(&translation); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("Couldn't read CSV line")
		}

		dictionary[i] = translation
		i++
	}

	c <- dictionary
}
