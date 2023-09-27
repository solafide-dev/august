package august

import (
	"fmt"
	"log"
	"reflect"
)

// an agust store is a map of id to interface{} that can be used to store any type of data
type AugustStore struct {
	parent *August
	name   string
	data   map[string]interface{}
}

func (as *AugustStore) Set(id string, val interface{}) error {
	(*as).data[id] = val
	return nil
}

func (as *AugustStore) Get(id string) (interface{}, error) {
	if val, ok := (*as).data[id]; ok {
		// we have the value
		storeType := reflect.ValueOf((*as).parent.storeRegistry[(*as).name]).Type()

		log.Printf("Found value for id: %s of type %T", id, storeType)

		return val, nil
	}

	return nil, fmt.Errorf("No value found for id: %s", id)
}
