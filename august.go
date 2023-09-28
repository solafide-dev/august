package august

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	l "log"
	"os"
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

var log *l.Logger

type AugustEventFunc func(event, store, id string)

type August struct {
	storeRegistry  map[string]reflect.Type // A map registrying the store types
	config         AugustConfig            // August configuration
	storage        map[string]AugustStore  // A map of all the stores
	eventFunc      AugustEventFunc         // A function to call when an event happens
	systemModCache []string                // Every time we modify a file, we add info about it so that FSNotify doesn't trigger on it
}

type AugustConfigOption string

func (c AugustConfigOption) String() string {
	return string(c)
}

const (
	// Storage directory for August to keep files.
	Config_StorageDir AugustConfigOption = "StorageDir"
	Config_Verbose    AugustConfigOption = "Verbose"
	Config_Format     AugustConfigOption = "Format"
	Config_FSNotify   AugustConfigOption = "FSNotify"
)

// AugustConfig stores basic configuration for August.
type AugustConfig struct {
	StorageDir string // Storage directory for August to keep files.
	Verbose    bool   // Enable logging.
	Format     string
	FSNotify   bool
}

// Create a new August instance.
func Init() *August {
	log = l.New(os.Stdout, "[August] ", l.LstdFlags|l.Lshortfile)
	log.SetOutput(io.Discard) // disable logging by default

	stores := make(map[string]reflect.Type)
	config := AugustConfig{
		StorageDir: "./storage",
		Verbose:    false,
		Format:     "json",
		FSNotify:   true,
	}
	storage := make(map[string]AugustStore)

	a := &August{
		storeRegistry:  stores,
		config:         config,
		storage:        storage,
		eventFunc:      func(event, store, id string) {},
		systemModCache: []string{},
	}

	return a
}

// Enable verbose logging.
func (a *August) Verbose() {
	log.SetOutput(os.Stdout)
}

// Set a config option.
func (a *August) Config(k AugustConfigOption, v interface{}) {
	log.Printf("Setting config: %s to %v", k, v)

	if k == Config_Verbose && v.(bool) {
		// set verbose mode if we configure that
		a.Verbose()
	}

	reflect.ValueOf(&a.config).Elem().FieldByName(k.String()).Set(reflect.ValueOf(v))
	log.Printf("Config: %+v", a.config)
}

func (a *August) SetEventFunc(f AugustEventFunc) {
	a.eventFunc = f
}

// Marshal an interface into the configured format.
func (a *August) Marshal(input interface{}) ([]byte, error) {
	switch a.config.Format {
	case "json":
		return json.MarshalIndent(input, "", "  ")
	case "yaml":
		return yaml.Marshal(input)
	case "xml":
		return xml.MarshalIndent(input, "", "  ")
	}
	return nil, fmt.Errorf("invalid format: %s", a.config.Format)
}

// Unmarshal an interface from the configured format.
func (a *August) Unmarshal(input []byte, output interface{}) error {
	switch a.config.Format {
	case "json":
		return json.Unmarshal(input, output)
	case "yaml":
		return yaml.Unmarshal(input, output)
	case "xml":
		return xml.Unmarshal(input, output)
	}
	return fmt.Errorf("invalid format: %s", a.config.Format)
}

// Get a store by name.
func (a *August) GetStore(name string) (*AugustStore, error) {
	if store, ok := a.storage[name]; ok {
		return &store, nil
	} else {
		return &AugustStore{}, fmt.Errorf("data store %s not found", name)
	}
}

// Register a store.
func (a *August) Register(name string, store interface{}) {
	log.Printf("Registering store: %s of type %T", name, store)

	a.storeRegistry[name] = reflect.TypeOf(store)
	a.storage[name] = AugustStore{
		name:   name,
		parent: a,
		data:   make(map[string]AugustStoreDataset),
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
		store.loadFromFile(id)
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

	if a.config.FSNotify {
		go func() {
			log.Println("Starting FSNotify watcher...")
			// We are going to watch the storage directory for changes
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.Fatal(err)
			}
			defer watcher.Close()

			// Start listening for events.
			go func() {
				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							return
						}
						eventType := event.Op.String()
						log.Println("event:", eventType, event.Name)
						if eventType == "WRITE" || eventType == "RENAME" || eventType == "CREATE" || eventType == "REMOVE" {
							// make sure names are normalized for windows systems
							file := strings.Replace(event.Name, `\\`, `/`, -1)
							file = strings.Replace(file, `\`, `/`, -1)

							// we need to parse the file changes to see what data needs to be updated

							storageDir := a.config.StorageDir + "/"
							if strings.HasPrefix(storageDir, "./") {
								storageDir = strings.Replace(storageDir, "./", "", 1)
							}

							nameAndId := strings.Replace(file, storageDir, "", 1)
							nameAndId = strings.Replace(nameAndId, "."+a.config.Format, "", 1)
							// now we should have group/id -- we need to split that
							parts := strings.Split(nameAndId, "/")
							if len(parts) != 2 {
								log.Println("[FS Notify] invalid file change event:", eventType, file)
								continue
							}

							storeName := parts[0]
							id := parts[1]
							method := "set"
							if eventType == "REMOVE" || eventType == "RENAME" {
								method = "delete"
							}

							if a.handleModCacheSkip(method, storeName, id) {
								continue
							}

							store, err := a.GetStore(storeName)
							if err != nil {
								log.Println("[FS Notify] error getting store:", err)
								continue
							}

							if eventType == "CREATE" || eventType == "WRITE" {
								// this should be treated as data being updates
								log.Printf("[FS Notify] File Modified: %s", file)
								err := store.loadFromFile(id)
								if err != nil {
									log.Println("error loading file:", err)
									continue
								}
							}

							if eventType == "REMOVE" || eventType == "RENAME" {
								// These both should be treated as data being deleted
								log.Printf("[FS Notify] File Deleted: %s", file)
								err := store.Delete(id)
								if err != nil {
									log.Println("error deleting file:", err)
									continue
								}
							}

						} else {
							log.Println("ignored file change event:", eventType, event.Name)
							continue
						}

					case err, ok := <-watcher.Errors:
						if !ok {
							return
						}
						log.Println("error:", err)
					}
				}
			}()

			for name, _ := range a.storeRegistry {
				// create directory for each store
				dir := a.config.StorageDir + "/" + name
				if err := watcher.Add(dir); err != nil {
					log.Fatal(err)
				}
			}

			// Block until an error occurs.
			err = <-watcher.Errors
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	return nil
}

// A weird function to check if the mod cache contains a string representing the action
// detected, and returns true + deletes the entry if it does.
func (a *August) handleModCacheSkip(method, name, id string) bool {
	cacheName := fmt.Sprintf("%s::%s::%s", method, name, id)
	for i, v := range a.systemModCache {
		if v == cacheName {
			log.Printf("[FS Notify] Found %s, skipping FS modify actions", cacheName)
			a.systemModCache = append(a.systemModCache[:i], a.systemModCache[i+1:]...)
			return true
		}
	}
	return false
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
