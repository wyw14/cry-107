package burner

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type Command struct {
	Identity     model.Identity `json:"identity"`
	FuelTPH      float64        `json:"fuel_tph"`
	PrimaryAir   float64        `json:"primary_air"`
	SecondaryAir float64        `json:"secondary_air"`
	IssuedAt     time.Time      `json:"issued_at"`
}

type Commander struct {
	mu      sync.RWMutex
	current Command
}

func NewCommander() *Commander {
	return &Commander{}
}

func (c *Commander) Issue(command Command) error {
	if !command.Identity.Valid() {
		return fmt.Errorf("burner command identity is invalid")
	}
	if command.FuelTPH < 0 || command.PrimaryAir < 0 || command.SecondaryAir < 0 {
		return fmt.Errorf("burner command targets cannot be negative")
	}
	command.IssuedAt = time.Now().UTC()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current.Identity.Generation > command.Identity.Generation {
		return fmt.Errorf("burner command is stale")
	}
	c.current = command
	return nil
}

func (c *Commander) Current() Command {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

func (c *Commander) Close(identity model.Identity) Command {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current.Identity.SameGeneration(identity) {
		c.current.FuelTPH = 0
		c.current.PrimaryAir = 0
		c.current.SecondaryAir = 0
		c.current.IssuedAt = time.Now().UTC()
	}
	return c.current
}
