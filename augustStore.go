package august

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"unicode"

	"github.com/google/uuid"
)

type AugustStoreDataset struct {
	data interface{} // The data we are storing
	lock *sync.Mutex // A lock to prevent concurrent writes to the data on disk
}

// an agust store represents individual data stores (folders) within the storage directory
type AugustStore struct {
	parent *August
	name   string
	data   map[string]AugustStoreDataset
}

func (as *AugustStore) set(id string, val interface{}) error {
	if err := as.ValidateId(id); err != nil {
		return err
	}

	if _, ok := (*as).data[id]; !ok {
		// create a new dataset
		(*as).data[id] = AugustStoreDataset{
			data: val,
			lock: &sync.Mutex{},
		}
	}

	dataSet := (*as).data[id]
	dataSet.data = val
	(*as).data[id] = dataSet

	(*as).parent.eventFunc("set", (*as).name, id)

	return nil
}

// Set stores a value in the store by id. Updating an existing value if it exists.
func (as *AugustStore) Set(id string, val interface{}) error {
	err := as.set(id, val)
	if err != nil {
		return err
	}

	err = as.SaveToFile(id)
	if err != nil {
		return err
	}

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
	if err := as.ValidateId(id); err != nil {
		return nil, err
	}

	if val, ok := (*as).data[id]; ok {
		// if the value is a pointer, we need to dereference it
		if reflect.TypeOf(val.data).Kind() == reflect.Ptr {
			return reflect.ValueOf(val.data).Elem().Interface(), nil
		}
		return val.data, nil
	}

	return nil, fmt.Errorf("no value found for id %s in store %s ", id, (*as).name)
}

// Get the lock from a dataset
func (as *AugustStore) getLock(id string) *sync.Mutex {
	if val, ok := (*as).data[id]; ok {
		// we have the value
		return val.lock
	}

	return &sync.Mutex{}
}

// Delete removes a value from the store by id.
func (as *AugustStore) Delete(id string) error {
	if err := as.ValidateId(id); err != nil {
		return err
	}

	if _, ok := (*as).data[id]; !ok {
		return fmt.Errorf("no value found for id %s in store %s ", id, (*as).name)
	}

	// remove the file
	filename := fmt.Sprintf("%s/%s/%s.%s", (*as).parent.config.StorageDir, (*as).name, id, (*as).parent.config.Format)
	log.Printf("Deleting file: %s", filename)
	if err := os.Remove(filename); err != nil {
		return err
	}

	// delete the value from the store
	delete((*as).data, id)

	(*as).parent.eventFunc("delete", (*as).name, id)

	return nil
}

// Get all the IDs in the store.
func (as *AugustStore) GetIds() []string {
	var ids []string
	for id := range (*as).data {
		ids = append(ids, id)
	}
	return ids
}

// GetAll returns all values in the store.
func (as *AugustStore) GetAll() (map[string]interface{}, error) {
	if len((*as).data) == 0 {
		return nil, fmt.Errorf("no data found for store: %s", (*as).name)
	}

	newSet := make(map[string]interface{})

	for id, val := range (*as).data {
		newSet[id] = val.data
	}

	return newSet, nil
}

// Purge will delete all of the data in a store.
func (as *AugustStore) Purge() error {
	for id := range (*as).data {
		if err := as.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

func (as *AugustStore) LoadFromFile(id string) error {
	if err := as.ValidateId(id); err != nil {
		return err
	}

	dataG := reflect.New(as.parent.storeRegistry[as.name]).Interface()
	data := dataG

	(*as).data[id] = AugustStoreDataset{
		data: data,
		lock: &sync.Mutex{},
	}

	filename := fmt.Sprintf("%s/%s/%s.%s", (*as).parent.config.StorageDir, (*as).name, id, (*as).parent.config.Format)
	log.Printf("Loading file: %s", filename)

	// read the file
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	data2 := as.data[id]
	data3 := data2.data

	// unmarshal the file
	if err := as.parent.Unmarshal(file, data3); err != nil {
		log.Printf("Error unmarshalling file: %s", err)
		return err
	}

	return nil
}

func (as *AugustStore) SaveToFile(id string) error {
	if err := as.ValidateId(id); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s/%s.%s", (*as).parent.config.StorageDir, (*as).name, id, (*as).parent.config.Format)
	log.Printf("Saving file: %s", filename)

	// get the value
	val, err := as.Get(id)
	if err != nil {
		return err
	}

	as.getLock(id).Lock()
	defer as.getLock(id).Unlock()

	// marshal the value
	data, err := (*as).parent.Marshal(val)
	if err != nil {
		return err
	}

	// write the file
	return os.WriteFile(filename, data, 0644)
}

func (as *AugustStore) ValidateId(id string) error {
	// make sure IDs only contain lower case letters and dashes

	if len(id) == 0 {
		return fmt.Errorf("id cannot be empty")
	}

	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-' {
			return fmt.Errorf("detected invalid character '%s' in id. Only alphanumeric chars and dashes are accepted", string(r))
		}
	}

	return nil
}
