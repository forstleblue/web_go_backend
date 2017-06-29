package utils

import (
	"errors"
	"log"
	"sort"
	"sync"
)

var counter int

//EventHandler stores parameters for a single event handler
type EventHandler struct {
	id       int
	once     bool
	Callback func(interface{}) interface{}
}

//******************

//EventHandlers stores all active EventHandler instances
type EventHandlers []EventHandler

//Implement the standard Go Sort Interface
func (eh EventHandlers) Len() int           { return len(eh) }
func (eh EventHandlers) Less(i, j int) bool { return eh[i].id < eh[j].id }
func (eh EventHandlers) Swap(i, j int)      { eh[i], eh[j] = eh[j], eh[i] }

//******************

//NewEventEmitter creates a new EventEmitter
func NewEventEmitter() *EventEmitter {
	e := EventEmitter{}
	e.events = make(map[string]*EventHandlers, 100)
	return &e
}

//******************

//EventEmitter manages event subscription and emission
type EventEmitter struct {
	sync.RWMutex
	events map[string]*EventHandlers
}

//On registers a new event handler for the given event name, returns total subscriber count
func (e *EventEmitter) On(eventname string, fn func(payload interface{}) interface{}) int {
	e.Lock()
	defer e.Unlock()

	//create handlers array for this event if it doesn't exist
	if e.events[eventname] == nil {
		e.events[eventname] = &EventHandlers{}
	}

	counter++
	eh := &EventHandler{id: counter, once: false, Callback: fn}
	*e.events[eventname] = append(*e.events[eventname], *eh)
	return counter
}

//Once registers a new event handler for the given event name to run once
func (e *EventEmitter) Once(eventname string, fn func(payload interface{}) interface{}) int {
	e.Lock()
	defer e.Unlock()
	if e.events[eventname] == nil {
		e.events[eventname] = &EventHandlers{}
	}

	counter++
	eh := &EventHandler{id: counter, once: true, Callback: fn}
	*e.events[eventname] = append(*e.events[eventname], *eh)

	return counter
}

//Off deregisters an event handler for the given event name
func (e *EventEmitter) Off(eventname string, id int) {
	e.Lock()
	defer e.Unlock()
	evts := *e.events[eventname]
	for i := range evts {
		if evts[i].id == id {
			log.Printf("REMOVING handler in event '%s' with id %d\n", eventname, id)
			*e.events[eventname] = append(evts[:i], evts[i+1:]...)
		}
	}
}

//OffAll deregisters all event handlers for the given eventname
func (e *EventEmitter) OffAll(eventname string) {
	e.Lock()
	defer e.Unlock()
	log.Println("REMOVING all event handlers for", eventname)
	e.events[eventname] = nil
}

//Reset deregisters all event handlers
func (e *EventEmitter) Reset() {
	e.Lock()
	defer e.Unlock()
	log.Println("REMOVING all event handlers")
	//e.events = nil
	e.events = make(map[string]*EventHandlers) //TODO try capped map and see if it's possible to auto-expand it

}

//Listeners returns all event handlers that are listening to an event
func (e *EventEmitter) Listeners(eventname string) EventHandlers {
	e.RLock()
	defer e.RUnlock()
	return *e.events[eventname]
}

//Emit emits an event to all the listening event handlers
func (e *EventEmitter) Emit(eventname string, msg interface{}, ordered bool) error {
	e.RLock()
	defer e.RUnlock()

	if e.events[eventname] == nil {
		// bail out if no event found
		return errors.New("No such event name found '" + eventname + "'")
	}
	handlers := *e.events[eventname]
	if ordered {
		sort.Sort(EventHandlers(handlers))
	}

	removeIfOnce := func(ev EventHandler) {
		if ev.once {
			log.Printf("REMOVING handler (specified once): in event '%s' with id %d\n", eventname, ev.id)
			for i := range handlers {
				if handlers[i].id == ev.id {
					*e.events[eventname] = append(handlers[:i], handlers[i+1:]...)
				}
			}
		}
	}
	//log.Println("handler count: ", len(handlers))
	for _, eh := range handlers {

		if &eh != nil {
			if ordered {
				eh.Callback(msg)
				removeIfOnce(eh)
			} else {
				go func() {
					eh.Callback(msg)
					removeIfOnce(eh)
				}()
			}
		}
	}

	return nil
}
