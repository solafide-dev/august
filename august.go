package august

import (
	"encoding/json"
	"fmt"
	"io"
	l "log"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

var log *l.Logger

type AugustEventFunc func(event, store, id string)

type August struct {
	storeRegistry map[string]interface{}
	config        AugustConfig
	storage       map[string]AugustStore
	eventFunc     AugustEventFunc
}

type AugustConfigOption string

func (c AugustConfigOption) String() string {
	return string(c)
}

const (
	// Storage directory for August to keep files.
	Config_StorageDir AugustConfigOption = "StorageDir"
)

// AugustConfig stores basic configuration for August.
type AugustConfig struct {
	StorageDir string // Storage directory for August to keep files.
	Verbose    bool   // Enable logging.
	Format     string
}

var defaultAugustConfig = AugustConfig{
	StorageDir: "./storage",
	Verbose:    false,
	Format:     "json",
}

// Create a new August instance.
func Init(c ...AugustConfig) *August {
	log = l.New(os.Stdout, "[August] ", l.LstdFlags|l.Lshortfile)

	stores := make(map[string]interface{})
	config := defaultAugustConfig
	storage := make(map[string]AugustStore)

	a := &August{
		storeRegistry: stores,
		config:        config,
		storage:       storage,
		eventFunc:     func(event, store, id string) {},
	}

	// if an AugustConfig is passed, override the default config with the set values
	if len(c) > 0 {

		values := reflect.ValueOf(c[0])
		//types := values.Type()
		for i := 0; i < values.NumField(); i++ {
			if values.Field(i).Interface() != nil && values.Field(i).Interface() != "" {
				// update a.config with the new value
				reflect.ValueOf(&a.config).Elem().Field(i).Set(values.Field(i))

			}
		}
	}

	if a.config.Verbose {
		log.Printf("August config: %+v", a.config)
	} else {
		log.SetOutput(io.Discard) // disable logging
	}

	return a
}

func (a *August) SetEventFunc(f AugustEventFunc) {
	a.eventFunc = f
}

func (a *August) Marshal(input interface{}) ([]byte, error) {
	switch a.config.Format {
	case "json":
		return json.MarshalIndent(input, "", "  ")
	case "yaml":
		return yaml.Marshal(input)
	}
	return nil, fmt.Errorf("invalid format: %s", a.config.Format)
}

func (a *August) Unmarshal(input []byte, output interface{}) error {
	switch a.config.Format {
	case "json":
		return json.Unmarshal(input, output)
	case "yaml":
		return yaml.Unmarshal(input, output)
	}
	return fmt.Errorf("invalid format: %s", a.config.Format)
}

func (a *August) GetStore(name string) (AugustStore, error) {
	if store, ok := a.storage[name]; ok {
		return store, nil
	} else {
		return AugustStore{}, fmt.Errorf("data store %s not found", name)
	}
}

// Register a store.
func (a *August) Register(name string, store interface{}) {
	log.Printf("Registering store: %s of type %T", name, store)
	ifame := reflect.TypeOf(store)
	a.storeRegistry[name] = ifame
	a.storage[name] = AugustStore{
		name:   name,
		parent: a,
		data:   make(map[string]interface{}),
	}
}

// Populate registry is used during initial startup to load any existing data.
func (a *August) populateRegistry(name string) error {
	if _, ok := a.storeRegistry[name]; !ok {
		return fmt.Errorf("store %s does not exists", name)
	}

	// check the directory for files and load them
	dir, err := os.ReadDir(a.config.StorageDir + "/" + name)
	if err != nil {
		return err
	}

	store := a.storage[name]

	for _, file := range dir {
		// skip invalid files
		if file.IsDir() || file.Type().IsRegular() && file.Name()[len(file.Name())-len(a.config.Format):] != a.config.Format {
			continue
		}

		id := file.Name()[:len(file.Name())-len(a.config.Format)-1]
		log.Printf("Loading file: %s for registry %s as ID %s", file.Name(), name, id)
		// read the file
		store.LoadFromFile(id)
	}
	return nil
}

// After august is configured, load data and monitor local files.
func (a *August) Run() error {
	// initialize storage structure
	if err := a.initStorage(); err != nil {
		return err
	}

	// populate registry
	for name, _ := range a.storeRegistry {
		if err := a.populateRegistry(name); err != nil {
			return err
		}
	}

	return nil
}

func (a *August) initStorage() error {
	// make initial storage directory
	// if it already exists, do nothing
	err := os.MkdirAll(a.config.StorageDir, os.ModePerm)
	if err != nil {
		return err
	}

	for name, _ := range a.storeRegistry {
		// create directory for each store
		dir := a.config.StorageDir + "/" + name
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
