package main

import (
	"log"

	"github.com/solafide-dev/august"
)

type Person struct {
	Name string `json:"name" yaml:"name" xml:"name"`
	Age  int    `json:"age" yaml:"age" xml:"age"`
}

type Group struct {
	Name    string   `json:"name" yaml:"name" xml:"name"`
	Members []Person `json:"members" yaml:"members" xml:"members"`
}

func main() {

	dataStore := august.Init(august.AugustConfig{
		Verbose:    true,        // enable logging
		StorageDir: "./storage", // set the storage directory
		Format:     "json",      // set the default format for data stores
	})

	dataStore.SetEventFunc(func(event, store, id string) {
		log.Printf("Storage Event Fired: %s, Store: %s, ID: %s", event, store, id)
	})

	// Set a config after the initial init is done
	dataStore.Register("people-group", Group{})

	// Initialize the data store (this initializes any registered data stores, and loads any existing data)
	if err := dataStore.Run(); err != nil {
		panic(err)
	}

	// Get a store
	peopleGroup, err := dataStore.GetStore("people-group")
	if err != nil {
		panic(err)
	}

	// Create a people group
	group := Group{
		Name: "Happy People",
		Members: []Person{
			{Name: "John Doe", Age: 30},
			{Name: "Jane Doe", Age: 28},
		},
	}

	// Set the group to the store as "happy-people"
	peopleGroup.Set("happy-people", group)

	// Get that group back
	p, err := peopleGroup.Get("happy-people")
	if err != nil {
		panic(err)
	}

	log.Printf("Group: %+v", p)

	// Generate a new people group with an auto-generated ID
	//newId, _ := peopleGroup.New(group)
	//p2, _ := peopleGroup.Get(newId)

	//log.Printf("Group2: %+v", p2)

	// Get all the defined IDS
	ids := peopleGroup.GetIds()

	for _, id := range ids {
		log.Printf("ID: %s", id)
		groupG, _ := peopleGroup.Get(id)
		group := groupG
		log.Printf("Group (%T): %+v", group, group)
	}

}
