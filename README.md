# August

August is a persistant data storage library that is based around folders and flat files (JSON, YAML, XML, ect).

Its main purpose is to provide a data store that prioritizes human readability and portability.

Its initial conception was to provide a data store for the [SimpleWorship](https://github.com/solafide-dev/simpleworship) software, since data sets won't be massive there, but human readability and portability are important.

## Usage

```go
// Initialize August
aug := august.Init()

// Set some configs (see below for available configs)
aug.Config(august.Config_Format, "json") // Set the format to yaml (default json)
aug.Config(august.Config_StorageDir, "./storage") // Set the storage directory (default ./storage)

// Setup the optional event fuction. This is useful if you need to subscribe to
// mutations in the data set to update other parts of your application.
aug.SetEventFunc(func(event, store, id string) {
    // Event will be one of create, update, delete
    // store will be the name of the store
    // id will be the id of the item that was created, updated, or deleted

    log.Printf("Event: %s, Store: %s, ID: %s", event, store, id)
})

type Person struct {
    Name string `json:"name"`
    Age int `json:"age"`
}

type Car struct {
    Make string `json:"make"`
    Model string `json:"model"`
}

// Register a data store to a type
aug.Register("people", Person{})
aug.Register("cars", Car{})

// Initialize the data store (this initializes any registered data stores, and loads any existing data)
if err := aug.Run(); err != nil {
    panic(err)
}

// Get a reference to the store
people, err := aug.GetStore("people")
if err != nil {
    panic(err)
}

// Add a person, with the ID "john-doe" to the store.
// The Set function will create / update the data store file.
err := people.Set("john-doe", Person{Name: "John Doe", Age: 30})

// Alternatively, you can allow august to generate a unique ID for you, if you don't want to manage them yourself.
id, err := people.New(Person{Name: "Jane Doe", Age: 28}) // ID will contain the new unique ID that was created.


// You can load data from a set using the Get function.
person, err := people.Get("john-doe") // or person, err := people.Get(id) to get Jane Doe we just created
fmt.Println(person.(Person).Name) // John Doe (notice the type assertion -- this is because the Get function returns an interface{})

// You can also optionally query all of the IDs in a set.
ids := people.GetIds()

for _, id := range ids {
    p, _ := people.Get(id)
    // Do something wich each data set item
}
```

This would result in a folder structure like this:

```
storage
├── people
│   ├── john-doe.json
│   └── f48929fa-8a8a-4af5-9511-a1c0794f9681.json
└── cars
```

You'll notice cars is empty, as we never created any data in that set.

As an example, the contents of `storage/people/john-doe.json` would be:
```json
{
    "name": "John Doe",
    "age": 30
}
```
If we have `Config_FSNotify` set to `true`, we can modify this file on disk and have the changes reflected in our datasets automatically.

```bash
# In Bash
jq '.age = 20' ./storage/people/john-doe.json | tee ./storage/people/john-doe.json
```

```go
// Back in our Go program
person, _ := people.Get("john-doe")
fmt.Println(person.(Person).Age) // 20
```


## Available Configs

| Config | Description | Default | Acceptable Values |
| --- | --- | --- | --- |
| Config_Format | The format to use for the data store files. | `json` | `json`, `yaml`, `xml` |
| Config_StorageDir | The directory to store the data store files in. | `./storage` | Any valid directory path |
| Config_FSNotify | Whether or not to use fsnotify to watch for changes to the data store files, and update the data store when they are modified | `true` | `true`, `false` |
| Config_Verbose | Whether or not to log verbose output (consider calling `*August.Verbose()` after init instead) | `false` | `true`, `false` |

## Why the name August?

At Sola Fide, we like for anything we make to echo our Christianity, as all the work we do is for Christ's Kingdom.

In the case of August, it is named after Saint Augustine of Hippo.

Saint Augustine was a philosopher and theologian who lived from 354 to 430 AD. He is considered one of the most important figures in the development of Western Christianity and was a major figure in bringing Christianity to dominance in the previously pagan Roman Empire.

![augustine-of-hippo](https://github.com/solafide-dev/august/assets/262524/f381eb80-7ec1-48b7-ad37-cdf1754a080d)
