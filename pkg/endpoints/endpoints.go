package endpoints

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// EndpointDB contains a set of endpoints
type EndpointDB struct {
	endpointsMutex sync.Mutex
	endpoints      map[string]string

	file string
}

// NewEndpointsDB gives an EndpointDB instance
func NewEndpointsDB(ctx context.Context, file string) *EndpointDB {
	db := &EndpointDB{
		endpoints: map[string]string{},
		file:      file,
	}
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}

	if err := watcher.Add(file); err != nil {
		fmt.Println("ERROR", err)
	}

	db.readFile()

	go func() {
		defer watcher.Close()

		for {
			select {
			// watch for events
			case <-watcher.Events:
				db.readFile()
			case err := <-watcher.Errors:
				if err != nil {
					panic(err)
				}
			}
		}

	}()
	return db
}

func (e *EndpointDB) readFile() {
	// open endpoint file
	file, err := os.Open(e.file)
	if err != nil {
		log.Println("error updating endpoints", err)
		return
	}
	defer file.Close()

	// read all lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// split line
		line := scanner.Text()
		// split line
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		// add endpoint
		e.endpointsMutex.Lock()
		e.endpoints[parts[0]] = parts[1]
		e.endpointsMutex.Unlock()
	}
}

func (e *EndpointDB) Get(endpoint string) (string, error) {
	e.endpointsMutex.Lock()
	defer e.endpointsMutex.Unlock()

	ok := false
	ep := ""
	for !ok {
		ep, ok = e.endpoints[endpoint]
		if !ok {
			// go one up the dns tree
			splitted := strings.Split(endpoint, ".")
			if len(splitted) < 2 {
				return "", fmt.Errorf("endpoint %s not found", endpoint)
			}
			endpoint = strings.Join(splitted[1:], ".")
		}
	}

	return ep, nil
}
