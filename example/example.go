package main

import (
	"log"
	"time"

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

	aug := august.Init()
	aug.Verbose()

	// aug.Config(august.Config_FSNotify, false) // Disable fsnotify (default true)
	// aug.Config(august.Config_Format, "yaml") // Set the format to yaml (default json)
	// aug.Config(august.Config_StorageDir, "./storage") // Set the storage directory (default ./storage)

	// Set an event function allows you to subscripe to events that happen in the data store
	// This is especially useful if you are using fsnotify, as your store may be mutated
	// without you realizing it by another process.
	aug.SetEventFunc(func(event, store, id string) {
		switch event {
		case "set":
			log.Printf("[DATA STORAGE] SET:%s:%s", store, id)
			store, _ := aug.GetStore(store)
			data, _ := store.Get(id)
			log.Printf("[DATA STORAGE] %s", data)
		case "delete":
			log.Printf("[DATA STORAGE] DELETE:%s:%s", store, id)

		}
	})

	// Set a config after the initial init is done
	aug.Register("people-group", Group{})

	// Initialize the data store (this initializes any registered data stores, and loads any existing data)
	if err := aug.Run(); err != nil {
		panic(err)
	}

	// Get a store
	peopleGroup, err := aug.GetStore("people-group")
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

	group2 := Group{
		Name: "Silly People",
		Members: []Person{
			{Name: "John Clown", Age: 30},
			{Name: "Jane Clown", Age: 28},
		},
	}

	// Generate a new people group with an auto-generated ID
	newId, _ := peopleGroup.New(group2)
	log.Printf("New ID: %s", newId)

	// Update Jane Clown to be 29
	janeClown, _ := peopleGroup.Get(newId)
	janeClown.(Group).Members[1].Age = 29

	peopleGroup.Set(newId, janeClown)

	// Get all the defined IDS
	ids := peopleGroup.GetIds()

	for _, id := range ids {
		log.Printf("ID: %s", id)
		groupG, _ := peopleGroup.Get(id)
		group := groupG
		log.Printf("Group (%T): %+v", group, group)
	}

	// Keep going forever so we can see the fsnotify events
	for {
		time.Sleep(1 * time.Second)
	}

}
