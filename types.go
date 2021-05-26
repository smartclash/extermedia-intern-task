package main

type Translation struct {
	English string
	French  string
}

type RepeatsCount struct {
	English		string
	French 		string
	Repetitions int
}

type Frequency struct {
	English 	string `csv:"English Word"`
	French 		string `csv:"French Word,omitempty"`
	Frequency 	int
}
