package august

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// an agust store represents individual data stores (folders) within the storage directory
type AugustStore struct {
	parent *August
	name   string
	data   map[string]interface{}
}

// Set stores a value in the store by id. Updating an existing value if it exists.
func (as *AugustStore) Set(id string, val interface{}) error {
	(*as).data[id] = val
	return nil
}

// New create a new value in the store, generating an ID for you and returning that ID.
func (as *AugustStore) New(val interface{}) (string, error) {
	// create a new ID using UUID
	id := uuid.New().String()
	err := as.Set(id, val)
	return id, err
}

// Get retrieves a value from the store by id.
func (as *AugustStore) Get(id string) (interface{}, error) {
	if val, ok := (*as).data[id]; ok {
		// we have the value
		storeType := reflect.ValueOf((*as).parent.storeRegistry[(*as).name]).Type()

		log.Printf("Found value for id: %s of type %T", id, storeType)

		return val, nil
	}

	return nil, fmt.Errorf("no value found for id %s in store %s ", id, (*as).name)
}

// Delete removes a value from the store by id.
func (as *AugustStore) Delete(id string) error {
	delete((*as).data, id)
	return nil
}

// GetAll returns all values in the store.
func (as *AugustStore) GetAll() (map[string]interface{}, error) {
	if len((*as).data) == 0 {
		return nil, fmt.Errorf("no data found for store: %s", (*as).name)
	}
	return (*as).data, nil
}
