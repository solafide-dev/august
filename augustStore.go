package august

import (
	"fmt"
	"os"
	"unicode"

	"github.com/google/uuid"
)

// an agust store represents individual data stores (folders) within the storage directory
type AugustStore struct {
	parent *August
	name   string
	data   map[string]interface{}
}

func (as *AugustStore) set(id string, val interface{}) error {
	if err := as.ValidateId(id); err != nil {
		return err
	}

	(*as).data[id] = val

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
		// we have the value
		return val, nil
	}

	return nil, fmt.Errorf("no value found for id %s in store %s ", id, (*as).name)
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

// GetAll returns all values in the store.
func (as *AugustStore) GetAll() (map[string]interface{}, error) {
	if len((*as).data) == 0 {
		return nil, fmt.Errorf("no data found for store: %s", (*as).name)
	}
	return (*as).data, nil
}

func (as *AugustStore) LoadFromFile(id string) error {
	if err := as.ValidateId(id); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s/%s.%s", (*as).parent.config.StorageDir, (*as).name, id, (*as).parent.config.Format)
	log.Printf("Loading file: %s", filename)

	// read the file
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var iface interface{}

	// unmarshal the file
	if err := (*as).parent.Unmarshal(file, &iface); err != nil {
		return err
	}

	return as.set(id, iface)
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
