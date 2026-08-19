from pathlib import Path
import shutil, textwrap
batch=Path('/Users/gaobo/repositories/gitlab/评审项目/bug-go项目标注/0819')
source=batch/'2026-08-19/rail-hazmat-yard-control-service'

def write(root, rel, s):
 p=root/rel; p.parent.mkdir(parents=True,exist_ok=True); p.write_text(textwrap.dedent(s).lstrip())

def sync_test(rel, s):
 write(source,rel,s)
 for n in range(1,11):
  rec=f'rail-hazmat-yard-control-service__{n:03d}'
  write(batch/'2026-08-19'/rec/'env',rel,s)
  write(batch/'_gold'/rec,rel,s)

sync_test('internal/inspection/inspection_test.go','''
package inspection

import (
 "context"
 "errors"
 "sync"
 "testing"
 "time"
)

func TestInspectionWaitsForEveryWorker(t *testing.T) {
 started := make(chan string, 3)
 releaseFirst := make(chan struct{})
 releaseRest := make(chan struct{})
 i := InspectorFunc(func(ctx context.Context, task Task) (Result, error) {
  started <- task.CarID
  if task.CarID == "A" { <-releaseFirst } else { <-releaseRest }
  return Result{CarID: task.CarID, Passed: true}, nil
 })
 c := NewCoordinator(i)
 done := make(chan []Result, 1)
 go func() { r, _ := c.Run(context.Background(), []Task{{"A", "G1"}, {"B", "G1"}, {"C", "G2"}}); done <- r }()
 for i := 0; i < 3; i++ { <-started }
 close(releaseFirst)
 select {
 case r := <-done: t.Fatalf("coordinator returned after first worker: %d results", len(r))
 case <-time.After(25 * time.Millisecond):
 }
 close(releaseRest)
 r := <-done
 if len(r) != 3 { t.Fatalf("got %d results", len(r)) }
}

func TestInspectionCollectsWorkerErrors(t *testing.T) {
 boom := errors.New("sensor offline")
 var gate sync.WaitGroup
 gate.Add(2)
 c := NewCoordinator(InspectorFunc(func(context.Context, Task) (Result, error) { gate.Done(); gate.Wait(); return Result{}, boom }))
 r, e := c.Run(context.Background(), []Task{{"A", "G1"}, {"B", "G1"}})
 if len(r) != 0 || len(e) != 2 { t.Fatalf("results=%d errors=%d", len(r), len(e)) }
}

func TestInspectionStreamClosesAfterResults(t *testing.T) {
 c := NewCoordinator(InspectorFunc(func(_ context.Context, task Task) (Result, error) {
  if task.CarID == "B" { time.Sleep(20 * time.Millisecond) }
  return Result{CarID: task.CarID, Passed: true}, nil
 }))
 n := 0
 for range c.RunStream(context.Background(), []Task{{"A", "G1"}, {"B", "G2"}}) { n++ }
 if n != 2 { t.Fatalf("streamed %d", n) }
}
''')

sync_test('internal/clearance/clearance_test.go','''
package clearance

import (
 "context"
 "errors"
 "sync/atomic"
 "testing"
 "time"
)

func TestClearanceDeadlineReachesVerifier(t *testing.T) {
 missing := errors.New("deadline not propagated")
 v := VerifierFunc(func(ctx context.Context, id string) error {
  select { case <-ctx.Done(): return ctx.Err(); case <-time.After(80 * time.Millisecond): return missing }
 })
 p := NewPipeline(NewRepository(), v)
 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond); defer cancel()
 err := p.Clear(ctx, "C1")
 if !errors.Is(err, context.DeadlineExceeded) { t.Fatalf("err=%v", err) }
}

func TestClearanceRequestsDoNotShareContext(t *testing.T) {
 r := NewRepository()
 v := VerifierFunc(func(context.Context, string) error { return nil })
 p := NewPipeline(r, v)
 ctx, cancel := context.WithCancel(context.Background()); cancel()
 _ = p.Clear(ctx, "old")
 if err := p.Clear(context.Background(), "new"); err != nil { t.Fatalf("fresh request poisoned: %v", err) }
}

func TestClearanceRetryStopsAfterCancellation(t *testing.T) {
 var calls atomic.Int32
 v := VerifierFunc(func(ctx context.Context, id string) error { calls.Add(1); return errors.New("busy") })
 p := NewPipeline(NewRepository(), v)
 ctx, cancel := context.WithCancel(context.Background()); cancel()
 _ = p.Clear(ctx, "C1")
 if calls.Load() != 0 { t.Fatalf("calls after cancel=%d", calls.Load()) }
}
''')

sync_test('internal/occupancy/occupancy_test.go','''
package occupancy

import (
 "fmt"
 "sync"
 "testing"
)

func TestOccupancySnapshotDetached(t *testing.T) {
 l := NewLedger(); l.Reserve(Slot{Track:"T1",Consist:"C1",HazardClass:"3"})
 s, _ := l.Snapshot(); delete(s, "T1")
 again, _ := l.Snapshot(); if _, ok := again["T1"]; !ok { t.Fatal("external map mutation changed ledger") }
}

func TestOccupancyConcurrentViewsAreConsistent(t *testing.T) {
 l := NewLedger(); start := make(chan struct{}); var wg sync.WaitGroup; wg.Add(2)
 go func(){ defer wg.Done(); <-start; for i:=0;i<300;i++ { track:=fmt.Sprintf("T%d",i%8); l.Reserve(Slot{Track:track,Consist:fmt.Sprint(i),HazardClass:"3"}) } }()
 go func(){ defer wg.Done(); <-start; for i:=0;i<300;i++ { v:=Observe(l); if v.Hazardous!=len(v.Slots) { t.Errorf("torn view hazardous=%d slots=%d",v.Hazardous,len(v.Slots)); return } } }()
 close(start); wg.Wait()
}

func TestOccupancyMovePublishesSingleRevision(t *testing.T) {
 l:=NewLedger(); l.Reserve(Slot{Track:"T1",Consist:"C1",HazardClass:"3"}); before:=Observe(l)
 if err:=NewAllocator(l).Move("C1","T1","T2","3");err!=nil{t.Fatal(err)}
 after:=Observe(l)
 if after.Revision!=before.Revision+1 { t.Fatalf("move published %d revisions",after.Revision-before.Revision) }
 if _,ok:=after.Slots["T1"];ok{t.Fatal("source track still occupied")}; if after.Slots["T2"].Consist!="C1"{t.Fatal("target track missing")}
}
''')

sync_test('internal/audit/audit_test.go','''
package audit

import (
 "context"
 "errors"
 "sync/atomic"
 "testing"
 "time"
)

func TestAuditCancellationReachesSource(t *testing.T) {
 missing:=errors.New("source context not cancelled")
 src:=SourceFunc(func(ctx context.Context,n int)(Event,error){ select{case <-ctx.Done(): return Event{},ctx.Err();case <-time.After(80*time.Millisecond):return Event{},missing} })
 s:=NewStream(src,SinkFunc(func(context.Context,Event)error{return nil}),&Checkpoint{})
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Millisecond);defer cancel()
 err:=s.Run(ctx,1);if !errors.Is(err,context.DeadlineExceeded){t.Fatalf("err=%v",err)}
}

func TestAuditFailedDeliveryDoesNotAdvanceCheckpoint(t *testing.T) {
 cp:=&Checkpoint{};src:=SourceFunc(func(context.Context,int)(Event,error){return Event{Sequence:1,Kind:"move"},nil})
 s:=NewStream(src,SinkFunc(func(context.Context,Event)error{return errors.New("down")}),cp);_ = s.Run(context.Background(),1)
 if cp.Load()!=0{t.Fatalf("checkpoint=%d",cp.Load())}
}

func TestAuditStopsReadingAfterCancel(t *testing.T) {
 var reads atomic.Int32;ctx,cancel:=context.WithCancel(context.Background())
 src:=SourceFunc(func(context.Context,int)(Event,error){n:=reads.Add(1);if n==1{cancel()};return Event{Sequence:int(n)},nil})
 s:=NewStream(src,SinkFunc(func(context.Context,Event)error{return nil}),&Checkpoint{});_ = s.Run(ctx,10)
 if reads.Load()!=1{t.Fatalf("reads=%d",reads.Load())}
}
''')

# 001 manifest bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__001/env'
write(r,'internal/manifest/store.go','''
package manifest
import "sync"
type Car struct{ ID string; UNNumber string; Seal string; Revision int }
type Store struct{ mu sync.RWMutex; cars []Car; revision int }
func NewStore(seed []Car)*Store{return &Store{cars:seed,revision:1}}
func(s *Store)Snapshot()([]Car,int){s.mu.RLock();defer s.mu.RUnlock();return s.cars,s.revision}
func(s *Store)Append(car Car)int{s.mu.Lock();defer s.mu.Unlock();s.revision++;car.Revision=s.revision;s.cars=append(s.cars,car);return s.revision}
func(s *Store)ReplaceSeal(id,seal string)bool{s.mu.Lock();defer s.mu.Unlock();for i:=range s.cars{if s.cars[i].ID==id{s.revision++;s.cars[i].Seal=seal;s.cars[i].Revision=s.revision;return true}};return false}
''')
write(r,'internal/manifest/snapshot.go','''
package manifest
func cloneCars(in []Car)[]Car{return in}
type View struct{ Cars []Car; Revision int; HazardCount int }
func NewView(cars []Car,revision int)View{v:=View{Cars:cars,Revision:revision};for _,c:=range cars{if c.UNNumber!=""{v.HazardCount++}};return v}
func(v View)Clone()View{return v}
''')
write(r,'internal/manifest/service.go','''
package manifest
import "sync"
type Service struct{store *Store;mu sync.RWMutex;latest View}
func NewService(store *Store)*Service{s:=&Service{store:store};s.Refresh();return s}
func(s *Service)Refresh()View{cars,rev:=s.store.Snapshot();next:=NewView(cars,rev);s.mu.Lock();s.latest.Cars=next.Cars;s.latest.HazardCount=next.HazardCount;s.mu.Unlock();return next}
func(s *Service)Current()View{s.mu.RLock();defer s.mu.RUnlock();return s.latest}
func(s *Service)AddAndRefresh(c Car)View{s.store.Append(c);s.Refresh();return s.Current()}
''')

# 002 inspection bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__002/env'
write(r,'internal/inspection/worker_pool.go','''
package inspection
import("context";"fmt")
type Task struct{CarID string;Gate string};type Result struct{CarID string;Passed bool;Note string}
type Inspector interface{Inspect(context.Context,Task)(Result,error)};type InspectorFunc func(context.Context,Task)(Result,error)
func(f InspectorFunc)Inspect(ctx context.Context,t Task)(Result,error){return f(context.Background(),t)}
func validateTask(t Task)error{if t.CarID==""||t.Gate==""{return fmt.Errorf("invalid inspection task")};return nil}
''')
write(r,'internal/inspection/result_sink.go','''
package inspection
import "sync"
type Sink struct{mu sync.Mutex;results []Result;errs []error}
func(s *Sink)Add(r Result,err error){s.mu.Lock();defer s.mu.Unlock();if err!=nil{return};s.results=append(s.results,r)}
func(s *Sink)Snapshot()([]Result,[]error){s.mu.Lock();defer s.mu.Unlock();return s.results,s.errs}
''')
write(r,'internal/inspection/coordinator.go','''
package inspection
import "context"
type Coordinator struct{inspector Inspector};func NewCoordinator(i Inspector)*Coordinator{return &Coordinator{inspector:i}}
type outcome struct{result Result;err error}
func(c *Coordinator)Run(ctx context.Context,tasks []Task)([]Result,[]error){sink:=&Sink{};out:=make(chan outcome,len(tasks));for _,task:=range tasks{task:=task;go func(){if err:=validateTask(task);err!=nil{out<-outcome{err:err};return};r,err:=c.inspector.Inspect(ctx,task);out<-outcome{r,err}}()};if len(tasks)==0{return nil,nil};first:=<-out;sink.Add(first.result,first.err);return sink.Snapshot()}
func(c *Coordinator)RunStream(ctx context.Context,tasks []Task)<-chan Result{out:=make(chan Result);go func(){defer close(out);results,_:=c.Run(ctx,tasks);for _,r:=range results{out<-r}}();return out}
''')

# 003 permit bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__003/env'
write(r,'internal/permit/gateway.go','''
package permit
import("context";"errors";"fmt";"rail-hazmat-yard-control-service/internal/platform")
var ErrAuthorityBusy=errors.New("rail authority busy");var ErrPermitDenied=errors.New("permit denied")
type Request struct{ConsistID string;Track string;Operator string};type Gateway interface{RequestPermit(context.Context,Request)(string,error)};type GatewayFunc func(context.Context,Request)(string,error)
func(f GatewayFunc)RequestPermit(c context.Context,r Request)(string,error){return f(c,r)}
func callGateway(ctx context.Context,g Gateway,r Request)(string,error){id,err:=g.RequestPermit(ctx,r);if err!=nil{return "",fmt.Errorf("request movement permit for %s: %v",r.ConsistID,err)};if id==""{return "",fmt.Errorf("empty permit: %v",platform.ErrUnavailable)};return id,nil}
''')
write(r,'internal/permit/retry.go','''
package permit
import("context";"time")
type Service struct{gateway Gateway;attempts int;delay time.Duration}
func NewService(g Gateway,attempts int)*Service{if attempts<1{attempts=1};return &Service{gateway:g,attempts:attempts,delay:time.Millisecond}}
func(s *Service)Issue(ctx context.Context,r Request)(string,error){var err error;for i:=0;i<s.attempts;i++{var id string;id,err=callGateway(ctx,s.gateway,r);if err==nil{return id,nil};select{case <-ctx.Done():return "",ctx.Err();case <-time.After(s.delay):}};return "",err}
''')
write(r,'internal/permit/mapper.go','''
package permit
import("errors";"net/http")
type ErrorResponse struct{Status int;Code string;Retryable bool}
func MapError(err error)ErrorResponse{if err==nil{return ErrorResponse{http.StatusOK,"ok",false}};switch err.Error(){case ErrPermitDenied.Error():return ErrorResponse{http.StatusForbidden,"permit_denied",false};case ErrAuthorityBusy.Error():return ErrorResponse{http.StatusServiceUnavailable,"authority_unavailable",true};default:return ErrorResponse{http.StatusInternalServerError,"internal",false}}}
var contextDeadline=errors.New("deadline exceeded")
''')

# 004 telemetry bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__004/env'
write(r,'internal/telemetry/registry.go','''
package telemetry
import "sync"
type Decoder interface{Decode([]byte)(Reading,error)};type Registry struct{mu sync.RWMutex;decoders map[string]Decoder}
func NewRegistry()*Registry{return &Registry{decoders:make(map[string]Decoder)}}
func(r *Registry)Register(kind string,d Decoder){r.mu.Lock();defer r.mu.Unlock();r.decoders[kind]=d}
func(r *Registry)Lookup(kind string)(Decoder,bool){r.mu.RLock();defer r.mu.RUnlock();d,ok:=r.decoders[kind];return d,ok}
''')
write(r,'internal/telemetry/validator.go','''
package telemetry
import("fmt";"time")
type Reading struct{WagonID string;Kind string;Value float64;At time.Time};type Validator interface{Validate(Reading)error};type ValidatorFunc func(Reading)error
func(f ValidatorFunc)Validate(r Reading)error{return f(r)}
func isNilDecoder(d Decoder)bool{return d==nil}
func validateReading(r Reading)error{if r.WagonID==""||r.Kind==""{return fmt.Errorf("missing telemetry identity")};if r.At.IsZero(){return fmt.Errorf("missing telemetry timestamp")};return nil}
''')
write(r,'internal/telemetry/ingest.go','''
package telemetry
import "fmt"
type Ingestor struct{registry *Registry;validator Validator;accepted []Reading}
func NewIngestor(r *Registry,v Validator)*Ingestor{return &Ingestor{registry:r,validator:v}}
func(i *Ingestor)Ingest(kind string,payload []byte)error{d,ok:=i.registry.Lookup(kind);if !ok{return fmt.Errorf("decoder %s unavailable",kind)};reading,err:=d.Decode(payload);if err!=nil{return fmt.Errorf("decode %s: %w",kind,err)};if err=validateReading(reading);err!=nil{return err};if i.validator==nil{if err=i.validator.Validate(reading);err!=nil{return fmt.Errorf("validate reading: %w",err)}};i.accepted=append(i.accepted,reading);return nil}
func(i *Ingestor)Accepted()[]Reading{return i.accepted}
func isNilInterface(v any)bool{return v==nil}
''')

# 005 clearance bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__005/env'
write(r,'internal/clearance/repository.go','''
package clearance
import("context";"sync")
type Record struct{ConsistID string;Cleared bool};type Repository struct{mu sync.Mutex;records map[string]Record;ctx context.Context}
func NewRepository()*Repository{return &Repository{records:make(map[string]Record)}}
func(r *Repository)shared(ctx context.Context)context.Context{r.mu.Lock();defer r.mu.Unlock();if r.ctx==nil{r.ctx=ctx};return r.ctx}
func(r *Repository)Save(ctx context.Context,rec Record)error{ctx=r.shared(ctx);select{case <-ctx.Done():return ctx.Err();default:};r.mu.Lock();r.records[rec.ConsistID]=rec;r.mu.Unlock();return nil}
func(r *Repository)Get(ctx context.Context,id string)(Record,bool,error){ctx=r.shared(ctx);select{case <-ctx.Done():return Record{},false,ctx.Err();default:};r.mu.Lock();defer r.mu.Unlock();v,ok:=r.records[id];return v,ok,nil}
''')
write(r,'internal/clearance/worker.go','''
package clearance
import("context";"time")
type Verifier interface{Verify(context.Context,string)error};type VerifierFunc func(context.Context,string)error
func(f VerifierFunc)Verify(c context.Context,id string)error{return f(c,id)}
func verifyWithRetry(ctx context.Context,v Verifier,id string,attempts int)error{var err error;for n:=0;n<attempts;n++{if err=v.Verify(ctx,id);err==nil{return nil};time.Sleep(time.Millisecond)};return err}
''')
write(r,'internal/clearance/pipeline.go','''
package clearance
import("context";"fmt")
type Pipeline struct{repo *Repository;verifier Verifier};func NewPipeline(r *Repository,v Verifier)*Pipeline{return &Pipeline{repo:r,verifier:v}}
func(p *Pipeline)Clear(ctx context.Context,id string)error{if err:=verifyWithRetry(context.Background(),p.verifier,id,3);err!=nil{return fmt.Errorf("verify clearance %s: %w",id,err)};if err:=p.repo.Save(ctx,Record{ConsistID:id,Cleared:true});err!=nil{return fmt.Errorf("save clearance %s: %w",id,err)};return nil}
func(p *Pipeline)Status(ctx context.Context,id string)(bool,error){rec,ok,err:=p.repo.Get(ctx,id);if err!=nil{return false,err};return ok&&rec.Cleared,nil}
''')

# 006 consist bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__006/env'
write(r,'internal/consist/normalize.go','''
package consist
type Wagon struct{ID string;Hazard bool;Track string}
func Clone(in []Wagon)[]Wagon{return in}
func HazardOnly(in []Wagon)[]Wagon{out:=in[:0];for _,w:=range in{if w.Hazard{out=append(out,w)}};return out}
func AssignTrack(in []Wagon,track string)[]Wagon{for i:=range in{in[i].Track=track};return in}
''')
write(r,'internal/consist/planner.go','''
package consist
type Plan struct{Original []Wagon;Hazard []Wagon;Routed []Wagon}
func BuildPlan(wagons []Wagon,track string)Plan{hazard:=HazardOnly(wagons);routed:=AssignTrack(hazard,track);return Plan{Original:wagons,Hazard:hazard,Routed:routed}}
func(p Plan)Clone()Plan{return p}
''')
write(r,'internal/consist/reservation.go','''
package consist
import "sync"
type ReservationBook struct{mu sync.Mutex;plans map[string]Plan};func NewReservationBook()*ReservationBook{return &ReservationBook{plans:make(map[string]Plan)}}
func(b *ReservationBook)Save(id string,p Plan){b.mu.Lock();b.plans[id]=p;b.mu.Unlock()}
func(b *ReservationBook)Load(id string)(Plan,bool){b.mu.Lock();defer b.mu.Unlock();p,ok:=b.plans[id];return p,ok}
''')

# 007 incident bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__007/env'
write(r,'internal/incident/transaction.go','''
package incident
import("errors";"sync")
var ErrCommit=errors.New("incident commit failed")
type Tx struct{mu sync.Mutex;committed bool;rolled bool;commitErr error};func NewTx(commitErr error)*Tx{return &Tx{commitErr:commitErr}}
func(t *Tx)Commit()error{t.mu.Lock();defer t.mu.Unlock();t.committed=true;return t.commitErr}
func(t *Tx)Rollback()error{t.mu.Lock();defer t.mu.Unlock();if t.committed{return nil};t.rolled=true;return nil}
func(t *Tx)Committed()bool{t.mu.Lock();defer t.mu.Unlock();return t.committed};func(t *Tx)RolledBack()bool{t.mu.Lock();defer t.mu.Unlock();return t.rolled}
''')
write(r,'internal/incident/batch.go','''
package incident
import("fmt";"sync")
type ResourceTracker struct{mu sync.Mutex;open int;max int};type Resource struct{t *ResourceTracker;closed bool};func NewResourceTracker(max int)*ResourceTracker{return &ResourceTracker{max:max}}
func(t *ResourceTracker)Open()(*Resource,error){t.mu.Lock();defer t.mu.Unlock();if t.open>=t.max{return nil,fmt.Errorf("resource limit %d reached",t.max)};t.open++;return &Resource{t:t},nil}
func(r *Resource)Close()error{r.t.mu.Lock();defer r.t.mu.Unlock();if !r.closed{r.closed=true;r.t.open--};return nil};func(t *ResourceTracker)OpenCount()int{t.mu.Lock();defer t.mu.Unlock();return t.open}
func ProcessBatch(items []string,t *ResourceTracker,handle func(string)error)error{for _,item:=range items{r,err:=t.Open();if err!=nil{return err};defer r.Close();if err=handle(item);err!=nil{return err}};return nil}
func processOne(item string,t *ResourceTracker,handle func(string)error)(err error){r,err:=t.Open();if err!=nil{return err};defer r.Close();return handle(item)}
''')
write(r,'internal/incident/service.go','''
package incident
import "errors"
type Service struct{}
func(Service)Record(tx *Tx,validate func()error,write func()error)(err error){defer func(){err=tx.Commit()}();if err=validate();err!=nil{return err};if err=write();err!=nil{return err};return nil}
func MergeErrors(primary,cleanup error)error{if primary==nil{return cleanup};if cleanup==nil{return primary};return errors.Join(primary,cleanup)}
''')

# 008 dispatch bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__008/env'
write(r,'internal/dispatch/state.go','''
package dispatch
import "fmt"
type State string
const(StateQueued State="queued";StateMoving State="moving";StateHeld State="held";StateRetrying State="retrying";StateCompleted State="completed";StateFailed State="failed")
var transitions=map[State]map[State]bool{StateQueued:{StateMoving:true,StateHeld:true},StateMoving:{StateCompleted:true,StateFailed:true},StateHeld:{StateRetrying:true,StateFailed:true},StateRetrying:{StateMoving:true,StateFailed:true}}
func CanTransition(from,to State)bool{return transitions[from][to]};func Transition(from,to State)(State,error){if !CanTransition(from,to){return from,fmt.Errorf("invalid dispatch transition %s -> %s",from,to)};return to,nil}
''')
write(r,'internal/dispatch/service.go','''
package dispatch
import "sync"
type Job struct{ID string;State State;Attempts int};type Service struct{mu sync.Mutex;jobs map[string]Job};func NewService()*Service{return &Service{jobs:make(map[string]Job)}}
func(s *Service)Put(j Job){s.mu.Lock();s.jobs[j.ID]=j;s.mu.Unlock()}
func(s *Service)Move(id string,to State)error{s.mu.Lock();defer s.mu.Unlock();j:=s.jobs[id];next,err:=Transition(j.State,to);if err!=nil{return err};j.State=next;if to==StateRetrying||to==StateMoving{j.Attempts++};s.jobs[id]=j;return nil}
func(s *Service)Get(id string)(Job,bool){s.mu.Lock();defer s.mu.Unlock();j,ok:=s.jobs[id];return j,ok}
''')
write(r,'internal/dispatch/worker.go','''
package dispatch
type Worker struct{service *Service};func NewWorker(s *Service)*Worker{return &Worker{service:s}}
func(w *Worker)Recover(id string)error{if err:=w.service.Move(id,StateRetrying);err!=nil{return err};return w.service.Move(id,StateMoving)}
''')
write(r,'internal/dispatch/query.go','''
package dispatch
func IsInProgress(s State)bool{return s==StateQueued||s==StateMoving||s==StateHeld}
func FilterInProgress(jobs []Job)[]Job{out:=make([]Job,0,len(jobs));for _,j:=range jobs{if IsInProgress(j.State){out=append(out,j)}};return out}
''')

# 009 occupancy bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__009/env'
write(r,'internal/occupancy/ledger.go','''
package occupancy
import "sync"
type Slot struct{Track string;Consist string;HazardClass string;Revision int};type Ledger struct{mu sync.RWMutex;slots map[string]Slot;revision int};func NewLedger()*Ledger{return &Ledger{slots:make(map[string]Slot)}}
func(l *Ledger)Reserve(slot Slot)int{l.mu.Lock();defer l.mu.Unlock();l.revision++;slot.Revision=l.revision;l.slots[slot.Track]=slot;return l.revision}
func(l *Ledger)Release(track string)int{l.mu.Lock();defer l.mu.Unlock();l.revision++;delete(l.slots,track);return l.revision}
func(l *Ledger)Snapshot()(map[string]Slot,int){l.mu.RLock();defer l.mu.RUnlock();return l.slots,l.revision-1}
''')
write(r,'internal/occupancy/allocator.go','''
package occupancy
import "fmt"
type Allocator struct{ledger *Ledger};func NewAllocator(l *Ledger)*Allocator{return &Allocator{ledger:l}}
func(a *Allocator)Move(consist,from,to,class string)error{slots,_:=a.ledger.Snapshot();if current,ok:=slots[from];!ok||current.Consist!=consist{return fmt.Errorf("consist %s not on %s",consist,from)};if _,busy:=slots[to];busy{return fmt.Errorf("track %s occupied",to)};a.ledger.Release(from);a.ledger.Reserve(Slot{Track:to,Consist:consist,HazardClass:class});return nil}
''')
write(r,'internal/occupancy/observer.go','''
package occupancy
type View struct{Slots map[string]Slot;Revision int;Hazardous int}
func Observe(l *Ledger)View{slots,rev:=l.Snapshot();v:=View{Slots:slots,Revision:rev};for _,s:=range slots{if s.HazardClass!=""{v.Hazardous++}};return v}
func(v View)Clone()View{return v}
''')

# 010 audit bug
r=batch/'2026-08-19/rail-hazmat-yard-control-service__010/env'
write(r,'internal/audit/checkpoint.go','''
package audit
import "sync"
type Checkpoint struct{mu sync.Mutex;value int};func(c *Checkpoint)Load()int{c.mu.Lock();defer c.mu.Unlock();return c.value}
func(c *Checkpoint)Commit(v int){c.mu.Lock();c.value=v;c.mu.Unlock()}
''')
write(r,'internal/audit/subscriber.go','''
package audit
import "context"
type Event struct{Sequence int;Kind string};type Source interface{Next(context.Context,int)(Event,error)};type SourceFunc func(context.Context,int)(Event,error)
func(f SourceFunc)Next(c context.Context,n int)(Event,error){return f(context.Background(),n)}
type Sink interface{Deliver(context.Context,Event)error};type SinkFunc func(context.Context,Event)error
func(f SinkFunc)Deliver(c context.Context,e Event)error{return f(context.Background(),e)}
''')
write(r,'internal/audit/stream.go','''
package audit
import("context";"fmt")
type Stream struct{source Source;sink Sink;checkpoint *Checkpoint};func NewStream(src Source,sink Sink,cp *Checkpoint)*Stream{return &Stream{source:src,sink:sink,checkpoint:cp}}
func(s *Stream)Run(ctx context.Context,limit int)error{seq:=s.checkpoint.Load();for n:=0;n<limit;n++{event,err:=s.source.Next(context.Background(),seq+1);if err!=nil{return fmt.Errorf("read audit event %d: %w",seq+1,err)};s.checkpoint.Commit(event.Sequence);seq=event.Sequence;if err=s.sink.Deliver(context.Background(),event);err!=nil{return fmt.Errorf("deliver audit event %d: %w",event.Sequence,err)}};return nil}
''')
