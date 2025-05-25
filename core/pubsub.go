package core // Changed package name

import (
	"log"
	"sync"
	"time"
)

// FaderUpdate is the message type for fader changes.
// It refers to *Fader, which is now defined in this 'core' package (in types.go).
type FaderUpdate *Fader

// Subscriber defines a subscriber channel that receives FaderUpdate messages.
// Channels should be buffered by the creator if desired.
type Subscriber chan FaderUpdate

// Publisher manages a set of subscribers and broadcasts FaderUpdate messages.
type Publisher struct {
	subscribers map[Subscriber]bool // Use map for easy add/remove; bool value is arbitrary
	mu          sync.RWMutex        // Mutex to protect concurrent access to subscribers
}

// NewPublisher creates and returns a new Publisher instance.
func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make(map[Subscriber]bool),
	}
}

// Subscribe adds a new subscriber to the Publisher's list.
func (p *Publisher) Subscribe(sub Subscriber) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.subscribers[sub]; ok {
		log.Printf("PubSub: Subscriber already registered. Total: %d", len(p.subscribers))
		return
	}
	p.subscribers[sub] = true
	log.Printf("PubSub: New subscriber registered. Total: %d", len(p.subscribers))
}

// Unsubscribe removes a subscriber from the Publisher's list and closes its channel.
func (p *Publisher) Unsubscribe(sub Subscriber) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.subscribers[sub]; ok {
		delete(p.subscribers, sub)
		close(sub)
		log.Printf("PubSub: Subscriber unregistered. Remaining: %d", len(p.subscribers))
	} else {
		log.Printf("PubSub: Attempted to unregister a non-existent subscriber.")
	}
}

// Publish sends a FaderUpdate message to all registered subscribers.
// It sends non-blockingly to prevent one slow subscriber from halting others.
func (p *Publisher) Publish(fader *Fader) { // fader is *core.Fader
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.subscribers) == 0 {
		return
	}

	updateMessage := FaderUpdate(fader)
	log.Printf("PubSub: Publishing update for Fader ID: %s. Level: %.1f, Muted: %t. %d subscribers.",
		fader.ID, fader.Level, fader.Muted, len(p.subscribers))

	for sub := range p.subscribers {
		go func(s Subscriber, msg FaderUpdate) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PubSub: Recovered from panic during send to subscriber: %v. Subscriber might have been closed.", r)
				}
			}()
			
			select {
			case s <- msg:
				log.Printf("PubSub: Message successfully sent to a subscriber for fader %s.", msg.ID)
			case <-time.After(100 * time.Millisecond):
				log.Printf("PubSub: Warning: Subscriber channel for fader %s timed out or is full. Message dropped for this subscriber.", msg.ID)
			}
		}(sub, updateMessage)
	}
}

// FaderEvents global variable was removed in a previous step.
// NewPublisher() is called by NewAppServer in appserver.go.
