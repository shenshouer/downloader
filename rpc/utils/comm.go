package main

type (
	Person struct {
		Name 	string
		Age 	int
	}

	Student struct {
		Person
		School 	string
		Class 	string
		Grade 	string
	}
)