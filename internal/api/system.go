package api

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/bearing"
	"github.com/wyw14/cry-107/internal/burner"
	"github.com/wyw14/cry-107/internal/bypass"
	"github.com/wyw14/cry-107/internal/coalmill"
	"github.com/wyw14/cry-107/internal/conveyor"
	"github.com/wyw14/cry-107/internal/cooler"
	"github.com/wyw14/cry-107/internal/crusher"
	"github.com/wyw14/cry-107/internal/emission"
	"github.com/wyw14/cry-107/internal/feeder"
	"github.com/wyw14/cry-107/internal/gas"
	"github.com/wyw14/cry-107/internal/grate"
	"github.com/wyw14/cry-107/internal/interlock"
	"github.com/wyw14/cry-107/internal/journal"
	"github.com/wyw14/cry-107/internal/kiln"
	"github.com/wyw14/cry-107/internal/lube"
	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/operation"
	"github.com/wyw14/cry-107/internal/pressure"
	"github.com/wyw14/cry-107/internal/rawmix"
	"github.com/wyw14/cry-107/internal/sampler"
	"github.com/wyw14/cry-107/internal/shell"
	"github.com/wyw14/cry-107/internal/wasteheat"
)

type System struct {
	mu                 sync.RWMutex
	Generations        *operation.Generations
	Operations         *operation.Manager
	Runtime            *kiln.Runtime
	Loads              *kiln.LoadCoordinator
	Burner             *burner.Service
	Damper             *pressure.Damper
	Draft              *kiln.DraftLoop
	Feed               *feeder.Service
	RawMix             *rawmix.Planner
	Shell              *shell.Scanner
	Lining             *kiln.LiningMonitor
	Lube               *lube.Coordinator
	Drive              *kiln.DrivePermit
	Cooler             *cooler.Service
	CoolingAllocation  *cooler.Allocator
	WasteHeat          *wasteheat.DemandService
	FanBudget          *grate.FanBudget
	Recovery           *cooler.RecoveryGraph
	Bypass             *bypass.Service
	BypassShutdown     *bypass.Shutdown
	Enrichment         *bypass.Enrichment
	Samples            *sampler.Aggregator
	Circulation        *kiln.Circulation
	Inerting           *coalmill.InertingService
	MillStartup        *coalmill.Startup
	Emissions          *emission.Receiver
	Interlocks         *interlock.Manager
	Journal            *journal.Store
	Snapshot           *journal.SnapshotStore
	Trips              *interlock.TripService
	Pumps              *lube.PumpBank
	BearingTemperature *bearing.TemperatureMonitor
	CO                 *gas.COMonitor
	Diagnostics        map[string]any
	current            model.Identity
}

type replayCollector struct {
	Kinds []string
}

func (r *replayCollector) Apply(event model.Event) error {
	r.Kinds = append(r.Kinds, event.Kind)
	return nil
}

func NewSystem(dataDirectory string) (*System, error) {
	store, err := journal.NewStore(filepath.Join(dataDirectory, "kilnguard-events.jsonl"))
	if err != nil {
		return nil, err
	}
	generations := operation.NewGenerations()
	manager := operation.NewManager(generations)
	identity := generations.Current()
	feedPermit := feeder.NewPermit()
	feedPermit.Grant(identity, "preheat sequence established")
	feedService := feeder.NewService(feedPermit)
	ratio := burner.NewRatioLoop(15.4, 87.5)
	commander := burner.NewCommander()
	burnerService := burner.NewService(ratio, commander)
	damper := pressure.NewDamper(70)
	runtime := kiln.NewRuntime()
	loadCoordinator := kiln.NewLoadCoordinator(runtime, damper, burnerService)
	controller := pressure.NewController(0.7, 0.12, 46)
	modes := pressure.NewModeService(controller, 46)
	draft := kiln.NewDraftLoop(modes, -45, 46)
	fanBudget := grate.NewFanBudget(120)
	coolingAllocation := cooler.NewAllocator(fanBudget)
	fan := grate.NewFan()
	coolingService := cooler.NewService(coolingAllocation, fan)
	wasteService := wasteheat.NewDemandService(fanBudget)
	transport := conveyor.NewStateService()
	crusherPermit := crusher.NewPermit()
	recovery := cooler.NewRecoveryGraph(transport, crusherPermit)
	flow := bearing.NewFlowObserver(18)
	lubeCoordinator := lube.NewCoordinator(flow, 2.4)
	drivePermit := kiln.NewDrivePermit()
	tracker := shell.NewHotspotTracker()
	scanner := shell.NewScanner(tracker)
	lining := kiln.NewLiningMonitor()
	bypassService := bypass.NewService()
	drainer := conveyor.NewDrainer(20, 5*time.Millisecond)
	gate := pressure.NewGate(true)
	bypassShutdown := bypass.NewShutdown(bypassService, drainer, gate)
	aggregator := sampler.NewAggregator()
	enrichment := bypass.NewEnrichment(aggregator)
	circulation := kiln.NewCirculation()
	interlocks := interlock.NewManager()
	millPermit := interlock.NewMillPermit()
	window := gas.NewOxygenWindow(8, 30*time.Second)
	inerting := coalmill.NewInertingService(window, millPermit, 900)
	compensator := operation.NewCompensator()
	coTrip := interlock.NewCOTrip(interlocks, millPermit, compensator)
	startup := coalmill.NewStartup(compensator, coTrip)
	temperature, err := bearing.NewTemperatureMonitor(72, 85)
	if err != nil {
		return nil, err
	}
	coMonitor, err := gas.NewCOMonitor(900, 1400)
	if err != nil {
		return nil, err
	}
	system := &System{
		Generations: generations, Operations: manager, Runtime: runtime, Loads: loadCoordinator,
		Burner: burnerService, Damper: damper, Draft: draft, Feed: feedService, RawMix: rawmix.NewPlanner(),
		Shell: scanner, Lining: lining, Lube: lubeCoordinator, Drive: drivePermit,
		Cooler: coolingService, CoolingAllocation: coolingAllocation, WasteHeat: wasteService,
		FanBudget: fanBudget, Recovery: recovery, Bypass: bypassService, BypassShutdown: bypassShutdown,
		Enrichment: enrichment, Samples: aggregator, Circulation: circulation, Inerting: inerting, MillStartup: startup,
		Emissions: emission.NewReceiver(emission.DefaultLimits()), Interlocks: interlocks,
		Journal: store, Snapshot: journal.NewSnapshotStore(filepath.Join(dataDirectory, "kilnguard-state.json")),
		Pumps: lube.NewPumpBank("thrust-main", "thrust-standby"), BearingTemperature: temperature,
		CO: coMonitor, Diagnostics: make(map[string]any), current: identity,
	}
	system.Trips = interlock.NewTripService(interlocks, store, system)
	if err := system.bootstrap(); err != nil {
		return nil, fmt.Errorf("bootstrap control state: %w", err)
	}
	return system, nil
}

func (s *System) bootstrap() error {
	identity := s.CurrentIdentity()
	target, err := s.RawMix.Set(identity, 210, 0.96, 2.45)
	if err != nil {
		return err
	}
	blend, err := rawmix.DefaultBlend(target)
	if err != nil {
		return err
	}
	if _, err := s.Feed.Apply(target); err != nil {
		return err
	}
	if _, err := s.Feed.Observe(208, identity); err != nil {
		return err
	}
	if _, err := s.Pumps.Start(identity, "thrust-main"); err != nil {
		return err
	}
	if _, err := s.Pumps.ObservePressure(identity, "thrust-main", 3.1); err != nil {
		return err
	}
	if _, err := s.Lube.Begin(identity, "thrust-main"); err != nil {
		return err
	}
	if _, err := s.Lube.ConfirmPressure(identity, 3.1); err != nil {
		return err
	}
	handover, err := s.Lube.ConfirmReturnFlow(identity, 24)
	if err != nil {
		return err
	}
	s.Drive.ApplyHandover(handover)
	temperature := s.BearingTemperature.Observe(58)
	spot := s.Shell.Observe(151, 386, 90*time.Second)
	s.Lining.ApplyHotspots([]shell.Hotspot{spot})
	s.Draft.EnterManual()
	if _, err := s.Draft.ManualMove(46); err != nil {
		return err
	}
	s.Draft.EnterAuto(-45)
	if _, err := s.Draft.Tick(); err != nil {
		return err
	}
	recovery := s.Recovery.Begin(identity, 0)
	recovery, err = s.Recovery.ApplySpeedProof(identity, 100)
	if err != nil {
		return err
	}
	inerting := s.Inerting.Begin()
	if _, err := s.Inerting.AddNitrogen(inerting.SessionID, 950); err != nil {
		return err
	}
	start := time.Now().UTC()
	if _, err := s.Inerting.ObserveOxygen(gas.OxygenSample{SessionID: inerting.SessionID, Percent: 7.6, Observed: start}); err != nil {
		return err
	}
	inerting, err = s.Inerting.ObserveOxygen(gas.OxygenSample{SessionID: inerting.SessionID, Percent: 7.4, Observed: start.Add(31 * time.Second)})
	if err != nil {
		return err
	}
	if err := s.MillStartup.Begin(identity, 35); err != nil {
		return err
	}
	s.MillStartup.AdvanceHotAir(18)
	residence := uuid.New()
	if _, err := s.Samples.UpdateGas(sampler.GasSample{BatchID: residence, Chlorine: 11, Observed: time.Now().UTC()}); err != nil {
		return err
	}
	if _, err := s.Samples.UpdateDust(sampler.DustSample{BatchID: residence, MassKG: 24, Chloride: 1.8, Observed: time.Now().UTC()}); err != nil {
		return err
	}
	enrichment, err := s.Enrichment.Calculate(residence)
	if err != nil {
		return err
	}
	if _, err := s.Circulation.ApplyBleed(residence, enrichment.Value); err != nil {
		return err
	}
	measurement := model.Measurement{SensorID: "kiln-shell-01", BatchID: uuid.New(), Value: 386, Unit: "C", Revision: 1, Observed: time.Now().UTC(), QualityOK: true}
	if err := measurement.Validate(); err != nil {
		return err
	}
	proof := model.Proof{Name: "lubrication", Operation: identity.Operation, Generation: identity.Generation, Satisfied: handover.Complete, Detail: "pressure and return flow", ObservedAt: time.Now().UTC()}
	events, err := s.Journal.ReadAll()
	if err != nil {
		return err
	}
	collector := &replayCollector{}
	replay, err := journal.Replay(events, collector)
	if err != nil {
		return err
	}
	s.Diagnostics = map[string]any{
		"blend": blend, "pumps": s.Pumps.List(), "bearing": temperature,
		"recovery": recovery, "inerting": inerting, "measurement": measurement,
		"lubrication_proof_ready": proof.Ready(identity), "replay": replay, "replayed_kinds": collector.Kinds,
	}
	if err := s.Snapshot.Save(map[string]any{"state": s.State(), "diagnostics": s.Diagnostics}); err != nil {
		return err
	}
	var restored map[string]any
	loaded, err := s.Snapshot.Load(&restored)
	if err != nil {
		return err
	}
	s.Diagnostics["snapshot_loaded"] = loaded && restored != nil
	return nil
}

func (s *System) CurrentIdentity() model.Identity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

func (s *System) BeginOperation(name string) (model.Identity, error) {
	record, err := s.Operations.Begin(name)
	if err != nil {
		return model.Identity{}, err
	}
	s.mu.Lock()
	s.current = record.Identity
	s.mu.Unlock()
	return record.Identity, nil
}

func (s *System) RequestLoad(target float64) (map[string]any, error) {
	identity, err := s.BeginOperation("kiln-load-adjustment")
	if err != nil {
		return nil, err
	}
	plan, damper, err := s.Loads.Request(identity, target)
	if err != nil {
		return nil, err
	}
	event, err := model.NewEvent("kiln.load.requested", identity, plan)
	if err == nil {
		err = s.Journal.Append(event)
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"operation": identity, "plan": plan, "damper": damper}, nil
}

func (s *System) ConfirmAir(actual float64) (kiln.RuntimeState, error) {
	identity := s.CurrentIdentity()
	state, err := s.Loads.ConfirmAir(identity, actual)
	if err != nil {
		return kiln.RuntimeState{}, err
	}
	_, err = s.Operations.Complete(identity, "combustion air established")
	return state, err
}

func (s *System) AllocateCooling(outletTemp, air float64) (cooler.State, error) {
	return s.Cooler.Adjust(s.CurrentIdentity(), outletTemp, air)
}

func (s *System) AllocateWasteHeat(air float64) (wasteheat.Demand, error) {
	return s.WasteHeat.Request(s.CurrentIdentity(), air)
}

func (s *System) ObserveEmission(co, nox, dust, chlorine float64) ([]string, error) {
	identity := s.CurrentIdentity()
	coState, err := s.CO.Observe(identity, co)
	if err != nil {
		return nil, err
	}
	breaches, err := s.Emissions.Receive(emission.Sample{
		Identity: s.CurrentIdentity(), CO: co, NOx: nox, Dust: dust,
		Chlorine: chlorine, ObservedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	if coState.HighHigh {
		if _, err := s.Trips.Execute(context.Background(), identity, "stack-co", "CO high-high terminal"); err != nil {
			return breaches, err
		}
		if _, err := s.Runtime.Transition(model.PhaseShutdown, "emission interlock"); err != nil {
			return breaches, err
		}
	}
	return breaches, nil
}

func (s *System) State() map[string]any {
	return map[string]any{
		"kiln": s.Runtime.Snapshot(), "load_control": s.Loads.State(),
		"cooler": s.Cooler.Snapshot(), "fan_budget": s.FanBudget.Snapshot(),
		"waste_heat": s.WasteHeat.Snapshot(), "draft": s.DraftState(),
		"feed": s.Feed.Snapshot(), "lining": s.Lining.Zones(), "drive": s.Drive.Snapshot(),
		"operations": s.Operations.List(), "circulation": s.Circulation.Snapshot(),
	}
}

func (s *System) DraftState() map[string]any {
	pressureValue, damper, mode := s.Draft.Snapshot()
	return map[string]any{"pressure": pressureValue, "damper": damper, "mode": mode}
}

func (s *System) RunBypassShutdown(ctx context.Context, kilograms float64) (bypass.ShutdownResult, error) {
	identity, err := s.BeginOperation("bypass-shutdown")
	if err != nil {
		return bypass.ShutdownResult{}, err
	}
	if _, err := s.Bypass.Start(identity, 35); err != nil {
		return bypass.ShutdownResult{}, err
	}
	if err := s.BypassShutdown.Prepare(identity, kilograms); err != nil {
		return bypass.ShutdownResult{}, err
	}
	return s.BypassShutdown.Run(ctx, identity)
}

func (s *System) ValidateReadyState() error {
	if !s.CurrentIdentity().Valid() {
		return fmt.Errorf("current operation identity is invalid")
	}
	if s.Runtime.Snapshot().Process.Phase == "" {
		return fmt.Errorf("kiln process state is empty")
	}
	return nil
}

func (s *System) StopFeed(_ context.Context, identity model.Identity) error {
	s.Feed.Stop(identity)
	return nil
}

func (s *System) CloseFuel(_ context.Context, identity model.Identity) error {
	s.Burner.Close(identity)
	return nil
}

func (s *System) HoldCooling(_ context.Context, identity model.Identity) error {
	_, err := s.Cooler.Adjust(identity, 145, 70)
	return err
}
