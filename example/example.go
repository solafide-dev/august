package main

import (
	"log"

	"github.com/solafide-dev/august"
)

type Person struct {
	Name string `json:"name" yaml:"name"`
	Age  int    `json:"age" yaml:"age"`
}

type Group struct {
	Name    string
	Members []Person
}

type Car struct {
	Make  string `json:"make" yaml:"make"`
	Model string `json:"model" yaml:"model"`
	Year  int    `json:"year" yaml:"year"`
}

func main() {

	dataStore := august.Init(august.AugustConfig{
		Verbose:    true,        // enable logging
		StorageDir: "./storage", // set the storage directory
	})

	dataStore.SetEventFunc(func(event, store, id string) {
		log.Printf("Event: %s, Store: %s, ID: %s", event, store, id)
	})

	// Set a config after the initial init is done
	dataStore.Register("people-group", &Group{})
	dataStore.Register("cars", &Car{})

	// Initialize the data store (this initializes any registered data stores, and loads any existing data)
	if err := dataStore.Run(); err != nil {
		panic(err)
	}

	// Get a store
	peopleGroup, err := dataStore.GetStore("people-group")
	if err != nil {
		panic(err)
	}

	group := Group{
		Name: "Happy People",
		Members: []Person{
			{Name: "John Doe", Age: 30},
		},
	}

	peopleGroup.Set("happy-people", group)

	// Get that person back
	pGet, err := peopleGroup.Get("happy-people")
	if err != nil {
		panic(err)
	}
	p := pGet.(Group)
	if err != nil {
		panic(err)
	}

	newId, _ := peopleGroup.New(group)
	p2Get, _ := peopleGroup.Get(newId)
	p2 := p2Get.(Group)

	log.Println("Group:", group)
	log.Printf("Person %s of type %T", p.Name, p)
	log.Printf("Person2 %s of type %T", p2.Name, p2)

}
