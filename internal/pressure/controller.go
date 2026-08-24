package pressure

import (
	"math"
	"sync"
)

type ControllerState struct {
	Setpoint float64 `json:"setpoint"`
	Integral float64 `json:"integral"`
	Output   float64 `json:"output"`
	LastPV   float64 `json:"last_pv"`
}

type Controller struct {
	mu    sync.Mutex
	state ControllerState
	kp    float64
	ki    float64
}

func NewController(kp, ki, initialOutput float64) *Controller {
	return &Controller{kp: kp, ki: ki, state: ControllerState{Output: initialOutput}}
}

func (c *Controller) TrackManual(output, processValue float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state.Output = clamp(output, 0, 100)
	c.state.LastPV = processValue
	// Tracking retains the observed state; EnterAuto performs the transfer calculation.
	c.state.Integral += (output - c.state.Output) * 0.05
}

func (c *Controller) InitializeBumpless(currentOutput, setpoint, processValue float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state.Setpoint = setpoint
	c.state.LastPV = processValue
	c.state.Output = clamp(currentOutput, 0, 100)
	c.state.Integral = c.state.Output - c.kp*(setpoint-processValue)
}

func (c *Controller) Tick(processValue float64) ControllerState {
	c.mu.Lock()
	defer c.mu.Unlock()
	errorValue := c.state.Setpoint - processValue
	c.state.Integral = clamp(c.state.Integral+c.ki*errorValue, -100, 100)
	c.state.Output = clamp(c.kp*errorValue+c.state.Integral, 0, 100)
	c.state.LastPV = processValue
	return c.state
}

func (c *Controller) Restore(state ControllerState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state.Output = clamp(state.Output, 0, 100)
	state.Integral = clamp(state.Integral, -100, 100)
	c.state = state
}

func (c *Controller) Snapshot() ControllerState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func clamp(value, low, high float64) float64 {
	return math.Max(low, math.Min(high, value))
}
