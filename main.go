package main

import (
	"encoding/csv"
	"fmt"
	"github.com/jszwec/csvutil"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {
	dictionaryChannel := make(chan []Translation)
	repeatedWordsChannel := make(chan []RepeatsCount)
	translatedChannel := make(chan string)
	repeatedCSVChannel := make(chan string)

	wg.Add(6)
	translatedText := ""

	go ReadDictionary(dictionaryChannel)

	file, err := os.Open("t8.shakespeare.txt")
	if err != nil {
		log.Fatal("Couldn't open shakespeare file")
	}
	defer file.Close()

	buff := make([]byte, 5000)
	for {
		totalRead, err := file.Read(buff)
		if err != io.EOF && err != nil {
			log.Fatal("Couldn't read shakespeare file into buffer")
		} else if err == io.EOF {
			break
		}

		translatedText += string(buff[:totalRead])
	}

	theDictionary := <-dictionaryChannel

	go TranslateShakespeare(theDictionary, translatedText, translatedChannel)
	go CountWordRepeats(theDictionary, translatedText, repeatedWordsChannel)
	go WriteFile(<-translatedChannel, "t8.shakespeare.translated.txt")
	go WriteRepeatsCSV(<-repeatedWordsChannel, repeatedCSVChannel)
	go WriteFile(<-repeatedCSVChannel, "frequency.csv")

	wg.Wait()
}

func TranslateShakespeare(dictionary []Translation, text string, c chan<- string) {
	fmt.Println("Translating Shakespeare play")

	var dictionaryPair []string
	lowerText := strings.ToLower(text)

	for _, translation := range dictionary {
		dictionaryPair = append(dictionaryPair, translation.English, translation.French)
	}

	text = strings.NewReplacer(dictionaryPair...).Replace(lowerText)

	fmt.Println("Translated Shakespeare play")
	c <- text
	wg.Done()
}

func CountWordRepeats(words []Translation, text string, c chan<- []RepeatsCount) {
	fmt.Println("Counting word repeats")

	var repeats []RepeatsCount
	lowerText := strings.ToLower(text)

	for _, word := range words {
		theCount := strings.Count(lowerText, word.English)
		repeats = append(repeats, RepeatsCount{
			English:        word.English,
			French: 		word.French,
			Repetitions: 	theCount,
		})
	}

	fmt.Println("Counted word repeats")
	c <- repeats
	wg.Done()
}

func WriteFile(text string, name string) {
	fmt.Println("Writing file", name)

	file, err := os.Create(name)
	if err != nil {
		log.Fatal("Couldn't create", name, "file")
	}
	defer file.Close()

	_, err = io.WriteString(file, text)
	if err != nil {
		log.Fatal("Couldn't write to", name, "file")
	}

	if err = file.Sync(); err != nil {
		log.Fatal("Couldn't write to", name, "file")
	}

	fmt.Println("Finished writing file", name)
	wg.Done()
}

func WriteRepeatsCSV(repeats []RepeatsCount, c chan<- string) {
	fmt.Println("Generating content for frequency.csv")

	var csvData []Frequency
	for _, repeat := range repeats {
		csvData = append(csvData, Frequency{
			English:   repeat.English,
			French:    repeat.French,
			Frequency: repeat.Repetitions,
		})
	}

	buff, err := csvutil.Marshal(csvData)
	if err != nil {
		log.Fatal("Couldn't generate CSV fields")
	}

	fmt.Println("Generated content for frequency.csv")
	c <- string(buff)
	wg.Done()
}

func ReadDictionary(c chan []Translation) {
	fmt.Println("Reading french_dictionary.csv")

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

	fmt.Println("Finished reading french_dictionary.csv")
	c <- dictionary
	wg.Done()
}
