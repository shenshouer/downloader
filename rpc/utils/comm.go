package comm

type (
	Person struct {
		No 		int
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