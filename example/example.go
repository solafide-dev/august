package main

import (
	"log"

	"github.com/solafide-dev/august"
)

type Person struct {
	Name string
	Age  int
}

type Group struct {
	Name    string
	Members []Person
}

type Car struct {
	Make  string
	Model string
	Year  int
}

func main() {

	dataStore := august.Init(august.AugustConfig{
		Verbose:    true,        // enable logging
		StorageDir: "./storage", // set the storage directory
	})

	// Set a config after the initial init is done
	dataStore.Register("people", &Group{})
	dataStore.Register("cars", &Car{})

	// Initialize the data store (this initializes any registered data stores, and loads any existing data)
	if err := dataStore.Run(); err != nil {
		panic(err)
	}

	// Get a store
	people, err := dataStore.GetStore("people")
	if err != nil {
		panic(err)
	}

	person := Person{
		Name: "John Doe",
		Age:  30,
	}

	people.Set("john-doe", person)

	// Get that person back
	pGet, err := people.Get("john-doe")
	p := pGet.(Person)
	if err != nil {
		panic(err)
	}

	log.Println("People:", people)
	log.Printf("Person %s of type %T", p.Name, p)

}
