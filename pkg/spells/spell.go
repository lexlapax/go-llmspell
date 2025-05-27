// ABOUTME: Defines the Spell type and spell management functionality
// ABOUTME: Provides a library of pre-written spells for common LLM tasks

package spells

import (
	"time"
)

// Spell represents a magical script for LLM interactions
type Spell struct {
	// Name is the spell's identifier
	Name string
	
	// Description explains what the spell does
	Description string
	
	// Engine specifies which script engine to use
	Engine string
	
	// Script contains the actual spell code
	Script string
	
	// Requirements lists any required variables or functions
	Requirements []string
	
	// Tags for categorizing spells
	Tags []string
	
	// CreatedAt timestamp
	CreatedAt time.Time
	
	// Version for tracking spell evolution
	Version string
}

// Library manages a collection of spells
type Library struct {
	spells map[string]*Spell
}

// NewLibrary creates a new spell library
func NewLibrary() *Library {
	return &Library{
		spells: make(map[string]*Spell),
	}
}

// Add adds a spell to the library
func (l *Library) Add(spell *Spell) error {
	l.spells[spell.Name] = spell
	return nil
}

// Get retrieves a spell by name
func (l *Library) Get(name string) (*Spell, bool) {
	spell, ok := l.spells[name]
	return spell, ok
}

// List returns all spells in the library
func (l *Library) List() []*Spell {
	result := make([]*Spell, 0, len(l.spells))
	for _, spell := range l.spells {
		result = append(result, spell)
	}
	return result
}

// FindByTag returns spells with a specific tag
func (l *Library) FindByTag(tag string) []*Spell {
	var result []*Spell
	for _, spell := range l.spells {
		for _, t := range spell.Tags {
			if t == tag {
				result = append(result, spell)
				break
			}
		}
	}
	return result
}