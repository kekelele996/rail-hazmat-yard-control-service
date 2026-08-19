from pathlib import Path
import textwrap, json
root=Path(__file__).resolve().parents[1]

def w(path, content):
    p=root/path; p.parent.mkdir(parents=True,exist_ok=True); p.write_text(textwrap.dedent(content).lstrip())

w('go.mod','''
module rail-hazmat-yard-control-service

go 1.23.0
''')
w('README.md','''
# Rail Hazmat Yard Control Service

A Go service for railway hazardous-material yard operations. It coordinates consist manifests, inspection crews, movement permits, clearance deadlines, telemetry ingestion, incident transactions, dispatch state, track occupancy, and audit event delivery.

## Structure

- `cmd/server`: HTTP service entrypoint.
- `internal/*`: domain modules and memory-backed operational services.
- `internal/policy`: reusable industry policy evaluators.
- `web`: a small operations status page served by the Go process.

## Run

```bash
go run ./cmd/server
```

The service listens on `PORT` (default `18080`). Health is available at `/health`; `/api/summary` returns an operational summary.

## Test

```bash
go test ./...
go test -race ./...
```

## Environment

- `PORT`: HTTP port, defaults to `18080`.
- `YARD_CODE`: logical yard identifier, defaults to `NORTH-TRANSFER`.
''')
w('runtime_smoke.json', json.dumps({"mode":"service","start":"go run ./cmd/server","workdir":".","env":{"PORT":"18080","YARD_CODE":"NORTH-TRANSFER"},"ready_url":"http://127.0.0.1:18080/health","ready_status":[200],"startup_timeout":30},ensure_ascii=False,indent=2)+'\n')
w('web/index.html','''
<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Rail Hazmat Yard</title><style>body{font:16px system-ui;background:#101820;color:#e9f0f4;margin:0}main{max-width:900px;margin:60px auto;padding:32px}.card{background:#172630;border:1px solid #2c4655;border-radius:16px;padding:24px}h1{color:#ffc857}code{color:#7dd3fc}</style></head><body><main><div class="card"><h1>Rail Hazmat Yard Control</h1><p>Operational service is online. Machine health: <code>/health</code></p><p>Summary endpoint: <code>/api/summary</code></p></div></main></body></html>
''')

w('cmd/server/main.go','''
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
)

type summary struct { Yard string `json:"yard"`; Status string `json:"status"`; Modules int `json:"modules"` }

func main() {
    port := os.Getenv("PORT"); if port == "" { port = "18080" }
    yard := os.Getenv("YARD_CODE"); if yard == "" { yard = "NORTH-TRANSFER" }
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Header().Set("content-type","application/json"); _,_ = w.Write([]byte(`{"status":"ok"}`)) })
    mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) { w.Header().Set("content-type","application/json"); _ = json.NewEncoder(w).Encode(summary{Yard:yard,Status:"operational",Modules:10}) })
    mux.Handle("/", http.FileServer(http.Dir("web")))
    addr := ":"+port; log.Printf("rail hazmat yard service listening on %s", addr)
    if err := http.ListenAndServe(addr,mux); err != nil { log.Fatal(fmt.Errorf("serve yard control: %w",err)) }
}
''')

# Shared platform/domain files.
w('internal/platform/clock.go','''
package platform
import ("sync"; "time")
type Clock interface{ Now() time.Time }
type RealClock struct{}
func (RealClock) Now() time.Time { return time.Now().UTC() }
type ManualClock struct{ mu sync.Mutex; now time.Time }
func NewManualClock(t time.Time)*ManualClock{return &ManualClock{now:t.UTC()}}
func(c *ManualClock)Now()time.Time{c.mu.Lock();defer c.mu.Unlock();return c.now}
func(c *ManualClock)Advance(d time.Duration){c.mu.Lock();c.now=c.now.Add(d);c.mu.Unlock()}
''')
w('internal/platform/identity.go','''
package platform
import("fmt";"strings";"sync/atomic")
type IDFactory struct{ prefix string; seq atomic.Uint64 }
func NewIDFactory(prefix string)*IDFactory{return &IDFactory{prefix:strings.ToUpper(strings.TrimSpace(prefix))}}
func(f *IDFactory)Next()string{n:=f.seq.Add(1);return fmt.Sprintf("%s-%08d",f.prefix,n)}
''')
w('internal/platform/errors.go','''
package platform
import "errors"
var ErrNotFound=errors.New("not found")
var ErrConflict=errors.New("state conflict")
var ErrUnavailable=errors.New("dependency unavailable")
var ErrUnauthorized=errors.New("unauthorized operation")
var ErrCancelled=errors.New("operation cancelled")
''')

# BUG 001 manifest snapshot isolation.
w('internal/manifest/store.go','''
package manifest
import "sync"
type Car struct{ ID string; UNNumber string; Seal string; Revision int }
type Store struct{ mu sync.RWMutex; cars []Car; revision int }
func NewStore(seed []Car)*Store{return &Store{cars:cloneCars(seed),revision:1}}
func(s *Store)Snapshot()([]Car,int){s.mu.RLock();defer s.mu.RUnlock();return cloneCars(s.cars),s.revision}
func(s *Store)Append(car Car)int{s.mu.Lock();defer s.mu.Unlock();s.revision++;car.Revision=s.revision;s.cars=append(s.cars,car);return s.revision}
func(s *Store)ReplaceSeal(id,seal string)bool{s.mu.Lock();defer s.mu.Unlock();next:=cloneCars(s.cars);for i:=range next{if next[i].ID==id{s.revision++;next[i].Seal=seal;next[i].Revision=s.revision;s.cars=next;return true}};return false}
''')
w('internal/manifest/snapshot.go','''
package manifest
func cloneCars(in []Car)[]Car{if in==nil{return nil};out:=make([]Car,len(in));copy(out,in);return out}
type View struct{ Cars []Car; Revision int; HazardCount int }
func NewView(cars []Car,revision int)View{v:=View{Cars:cloneCars(cars),Revision:revision};for _,c:=range cars{if c.UNNumber!=""{v.HazardCount++}};return v}
func(v View)Clone()View{v.Cars=cloneCars(v.Cars);return v}
''')
w('internal/manifest/service.go','''
package manifest
import "sync"
type Service struct{ store *Store; mu sync.RWMutex; latest View }
func NewService(store *Store)*Service{s:=&Service{store:store};s.Refresh();return s}
func(s *Service)Refresh()View{cars,rev:=s.store.Snapshot();next:=NewView(cars,rev);s.mu.Lock();s.latest=next;s.mu.Unlock();return next.Clone()}
func(s *Service)Current()View{s.mu.RLock();defer s.mu.RUnlock();return s.latest.Clone()}
func(s *Service)AddAndRefresh(c Car)View{s.store.Append(c);return s.Refresh()}
''')
w('internal/manifest/manifest_test.go','''
package manifest
import("fmt";"sync";"testing")
func TestManifestSnapshotIsolation(t *testing.T){s:=NewStore([]Car{{ID:"A",UNNumber:"1203",Seal:"S1"}});snap,_:=s.Snapshot();if !s.ReplaceSeal("A","S2"){t.Fatal("replace failed")};if snap[0].Seal!="S1"{t.Fatalf("historical snapshot changed: %s",snap[0].Seal)}}
func TestManifestConcurrentViewStable(t *testing.T){s:=NewStore([]Car{{ID:"A",UNNumber:"1203",Seal:"S1"}});svc:=NewService(s);start:=make(chan struct{});var wg sync.WaitGroup;wg.Add(2);go func(){defer wg.Done();<-start;for i:=0;i<200;i++{svc.AddAndRefresh(Car{ID:fmt.Sprintf("C%d",i),UNNumber:"1993"})}}();go func(){defer wg.Done();<-start;for i:=0;i<200;i++{v:=svc.Current();if len(v.Cars)>0&&v.HazardCount!=len(v.Cars){t.Errorf("torn view hazards=%d cars=%d",v.HazardCount,len(v.Cars));return}}}();close(start);wg.Wait()}
func TestManifestRevisionMatchesPublishedView(t *testing.T){s:=NewStore(nil);svc:=NewService(s);for i:=0;i<25;i++{v:=svc.AddAndRefresh(Car{ID:fmt.Sprint(i),UNNumber:"1017"});if v.Revision!=len(v.Cars)+1{t.Fatalf("revision %d cars %d",v.Revision,len(v.Cars))}}}
''')

# BUG 002 inspection coordination.
w('internal/inspection/worker_pool.go','''
package inspection
import("context";"fmt")
type Task struct{ CarID string; Gate string }
type Result struct{ CarID string; Passed bool; Note string }
type Inspector interface{ Inspect(context.Context,Task)(Result,error) }
type InspectorFunc func(context.Context,Task)(Result,error)
func(f InspectorFunc)Inspect(ctx context.Context,t Task)(Result,error){return f(ctx,t)}
func validateTask(t Task)error{if t.CarID==""||t.Gate==""{return fmt.Errorf("invalid inspection task")};return nil}
''')
w('internal/inspection/result_sink.go','''
package inspection
import "sync"
type Sink struct{ mu sync.Mutex; results []Result; errs []error }
func(s *Sink)Add(r Result,err error){s.mu.Lock();defer s.mu.Unlock();if err!=nil{s.errs=append(s.errs,err);return};s.results=append(s.results,r)}
func(s *Sink)Snapshot()([]Result,[]error){s.mu.Lock();defer s.mu.Unlock();rr:=append([]Result(nil),s.results...);ee:=append([]error(nil),s.errs...);return rr,ee}
''')
w('internal/inspection/coordinator.go','''
package inspection
import("context";"sync")
type Coordinator struct{ inspector Inspector }
func NewCoordinator(i Inspector)*Coordinator{return &Coordinator{inspector:i}}
func(c *Coordinator)Run(ctx context.Context,tasks []Task)([]Result,[]error){sink:=&Sink{};var wg sync.WaitGroup;wg.Add(len(tasks));for _,task:=range tasks{task:=task;go func(){defer wg.Done();if err:=validateTask(task);err!=nil{sink.Add(Result{},err);return};r,err:=c.inspector.Inspect(ctx,task);sink.Add(r,err)}()};wg.Wait();return sink.Snapshot()}
func(c *Coordinator)RunStream(ctx context.Context,tasks []Task)<-chan Result{out:=make(chan Result);go func(){defer close(out);results,_:=c.Run(ctx,tasks);for _,r:=range results{select{case out<-r:case <-ctx.Done():return}}}();return out}
''')
w('internal/inspection/inspection_test.go','''
package inspection
import("context";"errors";"sync";"testing")
func TestInspectionWaitsForEveryWorker(t *testing.T){start:=make(chan struct{});release:=make(chan struct{});var once sync.Once;i:=InspectorFunc(func(ctx context.Context,task Task)(Result,error){once.Do(func(){close(start)});<-release;return Result{CarID:task.CarID,Passed:true},nil});c:=NewCoordinator(i);done:=make(chan []Result,1);go func(){r,_:=c.Run(context.Background(),[]Task{{"A","G1"},{"B","G1"},{"C","G2"}});done<-r}();<-start;close(release);r:=<-done;if len(r)!=3{t.Fatalf("got %d results",len(r))}}
func TestInspectionCollectsWorkerErrors(t *testing.T){boom:=errors.New("sensor offline");c:=NewCoordinator(InspectorFunc(func(context.Context,Task)(Result,error){return Result{},boom}));r,e:=c.Run(context.Background(),[]Task{{"A","G1"},{"B","G1"}});if len(r)!=0||len(e)!=2{t.Fatalf("results=%d errors=%d",len(r),len(e))}}
func TestInspectionStreamClosesAfterResults(t *testing.T){c:=NewCoordinator(InspectorFunc(func(_ context.Context,task Task)(Result,error){return Result{CarID:task.CarID,Passed:true},nil}));n:=0;for range c.RunStream(context.Background(),[]Task{{"A","G1"},{"B","G2"}}){n++};if n!=2{t.Fatalf("streamed %d",n)}}
''')

# BUG 003 permit error chain.
w('internal/permit/gateway.go','''
package permit
import("context";"errors";"fmt";"rail-hazmat-yard-control-service/internal/platform")
var ErrAuthorityBusy=errors.New("rail authority busy")
var ErrPermitDenied=errors.New("permit denied")
type Request struct{ ConsistID string; Track string; Operator string }
type Gateway interface{ RequestPermit(context.Context,Request)(string,error) }
type GatewayFunc func(context.Context,Request)(string,error)
func(f GatewayFunc)RequestPermit(c context.Context,r Request)(string,error){return f(c,r)}
func callGateway(ctx context.Context,g Gateway,r Request)(string,error){id,err:=g.RequestPermit(ctx,r);if err!=nil{return "",fmt.Errorf("request movement permit for %s: %w",r.ConsistID,err)};if id==""{return "",fmt.Errorf("empty permit: %w",platform.ErrUnavailable)};return id,nil}
''')
w('internal/permit/retry.go','''
package permit
import("context";"errors";"time";"rail-hazmat-yard-control-service/internal/platform")
type Service struct{ gateway Gateway; attempts int; delay time.Duration }
func NewService(g Gateway,attempts int)*Service{if attempts<1{attempts=1};return &Service{gateway:g,attempts:attempts,delay:time.Millisecond}}
func(s *Service)Issue(ctx context.Context,r Request)(string,error){var err error;for i:=0;i<s.attempts;i++{var id string;id,err=callGateway(ctx,s.gateway,r);if err==nil{return id,nil};if errors.Is(err,ErrPermitDenied)||errors.Is(err,platform.ErrUnauthorized){return "",err};if !errors.Is(err,ErrAuthorityBusy)&&!errors.Is(err,platform.ErrUnavailable){return "",err};select{case <-ctx.Done():return "",ctx.Err();case <-time.After(s.delay):}};return "",err}
''')
w('internal/permit/mapper.go','''
package permit
import("errors";"net/http";"rail-hazmat-yard-control-service/internal/platform")
type ErrorResponse struct{ Status int; Code string; Retryable bool }
func MapError(err error)ErrorResponse{switch{case errors.Is(err,ErrPermitDenied),errors.Is(err,platform.ErrUnauthorized):return ErrorResponse{http.StatusForbidden,"permit_denied",false};case errors.Is(err,ErrAuthorityBusy),errors.Is(err,platform.ErrUnavailable):return ErrorResponse{http.StatusServiceUnavailable,"authority_unavailable",true};case errors.Is(err,contextDeadline):return ErrorResponse{http.StatusGatewayTimeout,"deadline",true};default:return ErrorResponse{http.StatusInternalServerError,"internal",false}}}
var contextDeadline=errors.New("deadline exceeded")
''')
w('internal/permit/permit_test.go','''
package permit
import("context";"errors";"testing";"rail-hazmat-yard-control-service/internal/platform")
func TestPermitPreservesSentinelChain(t *testing.T){g:=GatewayFunc(func(context.Context,Request)(string,error){return "",ErrPermitDenied});_,err:=NewService(g,3).Issue(context.Background(),Request{ConsistID:"C1"});if !errors.Is(err,ErrPermitDenied){t.Fatalf("lost denial: %v",err)}}
func TestPermitDoesNotRetryUnauthorized(t *testing.T){calls:=0;g:=GatewayFunc(func(context.Context,Request)(string,error){calls++;return "",platform.ErrUnauthorized});_,_ = NewService(g,4).Issue(context.Background(),Request{ConsistID:"C1"});if calls!=1{t.Fatalf("unauthorized retried %d times",calls)}}
func TestPermitMapsBusyDependency(t *testing.T){g:=GatewayFunc(func(context.Context,Request)(string,error){return "",ErrAuthorityBusy});_,err:=NewService(g,1).Issue(context.Background(),Request{ConsistID:"C1"});r:=MapError(err);if r.Status!=503||!r.Retryable{t.Fatalf("mapping=%+v err=%v",r,err)}}
''')

# BUG 004 telemetry nil / typed nil.
w('internal/telemetry/registry.go','''
package telemetry
import "sync"
type Decoder interface{ Decode([]byte)(Reading,error) }
type Registry struct{ mu sync.RWMutex; decoders map[string]Decoder }
func NewRegistry()*Registry{return &Registry{decoders:make(map[string]Decoder)}}
func(r *Registry)Register(kind string,d Decoder){r.mu.Lock();defer r.mu.Unlock();if r.decoders==nil{r.decoders=make(map[string]Decoder)};r.decoders[kind]=d}
func(r *Registry)Lookup(kind string)(Decoder,bool){r.mu.RLock();defer r.mu.RUnlock();d,ok:=r.decoders[kind];return d,ok&&!isNilDecoder(d)}
''')
w('internal/telemetry/validator.go','''
package telemetry
import("fmt";"reflect";"time")
type Reading struct{ WagonID string; Kind string; Value float64; At time.Time }
type Validator interface{ Validate(Reading)error }
type ValidatorFunc func(Reading)error
func(f ValidatorFunc)Validate(r Reading)error{return f(r)}
func isNilDecoder(d Decoder)bool{if d==nil{return true};v:=reflect.ValueOf(d);switch v.Kind(){case reflect.Pointer,reflect.Map,reflect.Func,reflect.Interface,reflect.Slice:return v.IsNil()};return false}
func validateReading(r Reading)error{if r.WagonID==""||r.Kind==""{return fmt.Errorf("missing telemetry identity")};if r.At.IsZero(){return fmt.Errorf("missing telemetry timestamp")};return nil}
''')
w('internal/telemetry/ingest.go','''
package telemetry
import("fmt";"reflect")
type Ingestor struct{ registry *Registry; validator Validator; accepted []Reading }
func NewIngestor(r *Registry,v Validator)*Ingestor{return &Ingestor{registry:r,validator:v}}
func(i *Ingestor)Ingest(kind string,payload []byte)error{d,ok:=i.registry.Lookup(kind);if !ok{return fmt.Errorf("decoder %s unavailable",kind)};reading,err:=d.Decode(payload);if err!=nil{return fmt.Errorf("decode %s: %w",kind,err)};if err=validateReading(reading);err!=nil{return err};if !isNilInterface(i.validator){if err=i.validator.Validate(reading);err!=nil{return fmt.Errorf("validate reading: %w",err)}};i.accepted=append(i.accepted,reading);return nil}
func(i *Ingestor)Accepted()[]Reading{return append([]Reading(nil),i.accepted...)}
func isNilInterface(v any)bool{if v==nil{return true};rv:=reflect.ValueOf(v);switch rv.Kind(){case reflect.Pointer,reflect.Map,reflect.Func,reflect.Interface,reflect.Slice:return rv.IsNil()};return false}
''')
w('internal/telemetry/telemetry_test.go','''
package telemetry
import("errors";"testing";"time")
type decoder struct{}
func(*decoder)Decode([]byte)(Reading,error){return Reading{WagonID:"W1",Kind:"temp",Value:12,At:time.Now()},nil}
type nilDecoder struct{}
func(*nilDecoder)Decode([]byte)(Reading,error){panic("typed nil decoder called")}
type nilValidator struct{}
func(*nilValidator)Validate(Reading)error{panic("typed nil validator called")}
func TestTelemetryZeroRegistryCanRegister(t *testing.T){var r Registry;r.Register("temp",&decoder{});if _,ok:=r.Lookup("temp");!ok{t.Fatal("decoder missing")}}
func TestTelemetryRejectsTypedNilDecoder(t *testing.T){r:=NewRegistry();var d *nilDecoder;r.Register("temp",d);if _,ok:=r.Lookup("temp");ok{t.Fatal("typed nil decoder exposed")}}
func TestTelemetryValidationCannotBeBypassed(t *testing.T){r:=NewRegistry();r.Register("temp",&decoder{});want:=errors.New("temperature blocked");i:=NewIngestor(r,ValidatorFunc(func(Reading)error{return want}));if err:=i.Ingest("temp",nil);!errors.Is(err,want){t.Fatalf("validation lost: %v",err)};if len(i.Accepted())!=0{t.Fatal("invalid reading accepted")}}
''')

# BUG 005 clearance context.
w('internal/clearance/repository.go','''
package clearance
import("context";"sync")
type Record struct{ ConsistID string; Cleared bool }
type Repository struct{ mu sync.Mutex; records map[string]Record }
func NewRepository()*Repository{return &Repository{records:make(map[string]Record)}}
func(r *Repository)Save(ctx context.Context,rec Record)error{select{case <-ctx.Done():return ctx.Err();default:};r.mu.Lock();r.records[rec.ConsistID]=rec;r.mu.Unlock();return nil}
func(r *Repository)Get(ctx context.Context,id string)(Record,bool,error){select{case <-ctx.Done():return Record{},false,ctx.Err();default:};r.mu.Lock();defer r.mu.Unlock();v,ok:=r.records[id];return v,ok,nil}
''')
w('internal/clearance/worker.go','''
package clearance
import("context";"time")
type Verifier interface{ Verify(context.Context,string)error }
type VerifierFunc func(context.Context,string)error
func(f VerifierFunc)Verify(c context.Context,id string)error{return f(c,id)}
func verifyWithRetry(ctx context.Context,v Verifier,id string,attempts int)error{var err error;for n:=0;n<attempts;n++{if err=ctx.Err();err!=nil{return err};if err=v.Verify(ctx,id);err==nil{return nil};select{case <-ctx.Done():return ctx.Err();case <-time.After(time.Millisecond):}};return err}
''')
w('internal/clearance/pipeline.go','''
package clearance
import("context";"fmt")
type Pipeline struct{ repo *Repository; verifier Verifier }
func NewPipeline(r *Repository,v Verifier)*Pipeline{return &Pipeline{repo:r,verifier:v}}
func(p *Pipeline)Clear(ctx context.Context,id string)error{if err:=ctx.Err();err!=nil{return err};if err:=verifyWithRetry(ctx,p.verifier,id,3);err!=nil{return fmt.Errorf("verify clearance %s: %w",id,err)};if err:=p.repo.Save(ctx,Record{ConsistID:id,Cleared:true});err!=nil{return fmt.Errorf("save clearance %s: %w",id,err)};return nil}
func(p *Pipeline)Status(ctx context.Context,id string)(bool,error){rec,ok,err:=p.repo.Get(ctx,id);if err!=nil{return false,err};return ok&&rec.Cleared,nil}
''')
w('internal/clearance/clearance_test.go','''
package clearance
import("context";"errors";"sync/atomic";"testing";"time")
func TestClearanceDeadlineReachesVerifier(t *testing.T){seen:=make(chan struct{});v:=VerifierFunc(func(ctx context.Context,id string)error{<-ctx.Done();close(seen);return ctx.Err()});p:=NewPipeline(NewRepository(),v);ctx,cancel:=context.WithTimeout(context.Background(),10*time.Millisecond);defer cancel();err:=p.Clear(ctx,"C1");if !errors.Is(err,context.DeadlineExceeded){t.Fatalf("err=%v",err)};select{case <-seen:case <-time.After(time.Second):t.Fatal("verifier did not see deadline")}}
func TestClearanceRequestsDoNotShareContext(t *testing.T){r:=NewRepository();v:=VerifierFunc(func(context.Context,string)error{return nil});p:=NewPipeline(r,v);ctx,cancel:=context.WithCancel(context.Background());cancel();_ = p.Clear(ctx,"old");if err:=p.Clear(context.Background(),"new");err!=nil{t.Fatalf("fresh request poisoned: %v",err)}}
func TestClearanceRetryStopsAfterCancellation(t *testing.T){var calls atomic.Int32;v:=VerifierFunc(func(ctx context.Context,id string)error{calls.Add(1);return errors.New("busy")});p:=NewPipeline(NewRepository(),v);ctx,cancel:=context.WithCancel(context.Background());cancel();_ = p.Clear(ctx,"C1");if calls.Load()!=0{t.Fatalf("calls after cancel=%d",calls.Load())}}
''')

# BUG 006 consist slice ownership.
w('internal/consist/normalize.go','''
package consist
type Wagon struct{ ID string; Hazard bool; Track string }
func Clone(in []Wagon)[]Wagon{if in==nil{return nil};out:=make([]Wagon,len(in));copy(out,in);return out}
func HazardOnly(in []Wagon)[]Wagon{out:=make([]Wagon,0,len(in));for _,w:=range in{if w.Hazard{out=append(out,w)}};return out}
func AssignTrack(in []Wagon,track string)[]Wagon{out:=Clone(in);for i:=range out{out[i].Track=track};return out}
''')
w('internal/consist/planner.go','''
package consist
type Plan struct{ Original []Wagon; Hazard []Wagon; Routed []Wagon }
func BuildPlan(wagons []Wagon,track string)Plan{original:=Clone(wagons);hazard:=HazardOnly(wagons);routed:=AssignTrack(hazard,track);return Plan{Original:original,Hazard:Clone(hazard),Routed:routed}}
func(p Plan)Clone()Plan{return Plan{Original:Clone(p.Original),Hazard:Clone(p.Hazard),Routed:Clone(p.Routed)}}
''')
w('internal/consist/reservation.go','''
package consist
import "sync"
type ReservationBook struct{ mu sync.Mutex; plans map[string]Plan }
func NewReservationBook()*ReservationBook{return &ReservationBook{plans:make(map[string]Plan)}}
func(b *ReservationBook)Save(id string,p Plan){b.mu.Lock();b.plans[id]=p.Clone();b.mu.Unlock()}
func(b *ReservationBook)Load(id string)(Plan,bool){b.mu.Lock();defer b.mu.Unlock();p,ok:=b.plans[id];return p.Clone(),ok}
''')
w('internal/consist/consist_test.go','''
package consist
import("reflect";"testing")
func TestConsistFilteringDoesNotMutateInput(t *testing.T){in:=[]Wagon{{"A",false,""},{"B",true,""},{"C",false,""}};before:=Clone(in);_ = HazardOnly(in);if !reflect.DeepEqual(in,before){t.Fatalf("input changed: %#v",in)}}
func TestConsistPlanSlicesAreIndependent(t *testing.T){p:=BuildPlan([]Wagon{{"A",true,""},{"B",true,""}},"T9");p.Routed[0].ID="changed";if p.Hazard[0].ID=="changed"||p.Original[0].ID=="changed"{t.Fatal("plan slices share storage")}}
func TestConsistReservationReturnsDetachedPlan(t *testing.T){b:=NewReservationBook();b.Save("P",BuildPlan([]Wagon{{"A",true,""}},"T1"));p,_:=b.Load("P");p.Routed[0].Track="BAD";again,_:=b.Load("P");if again.Routed[0].Track!="T1"{t.Fatalf("stored plan polluted: %#v",again)}}
''')

# BUG 007 incident resources/defer.
w('internal/incident/transaction.go','''
package incident
import("errors";"sync")
var ErrCommit=errors.New("incident commit failed")
type Tx struct{ mu sync.Mutex; committed bool; rolled bool; commitErr error }
func NewTx(commitErr error)*Tx{return &Tx{commitErr:commitErr}}
func(t *Tx)Commit()error{t.mu.Lock();defer t.mu.Unlock();if t.commitErr!=nil{return t.commitErr};t.committed=true;return nil}
func(t *Tx)Rollback()error{t.mu.Lock();t.rolled=true;t.mu.Unlock();return nil}
func(t *Tx)Committed()bool{t.mu.Lock();defer t.mu.Unlock();return t.committed}
func(t *Tx)RolledBack()bool{t.mu.Lock();defer t.mu.Unlock();return t.rolled}
''')
w('internal/incident/batch.go','''
package incident
import("fmt";"sync")
type ResourceTracker struct{ mu sync.Mutex; open int; max int }
type Resource struct{ t *ResourceTracker; closed bool }
func NewResourceTracker(max int)*ResourceTracker{return &ResourceTracker{max:max}}
func(t *ResourceTracker)Open()(*Resource,error){t.mu.Lock();defer t.mu.Unlock();if t.open>=t.max{return nil,fmt.Errorf("resource limit %d reached",t.max)};t.open++;return &Resource{t:t},nil}
func(r *Resource)Close()error{r.t.mu.Lock();defer r.t.mu.Unlock();if !r.closed{r.closed=true;r.t.open--};return nil}
func(t *ResourceTracker)OpenCount()int{t.mu.Lock();defer t.mu.Unlock();return t.open}
func ProcessBatch(items []string,t *ResourceTracker,handle func(string)error)error{for _,item:=range items{if err:=processOne(item,t,handle);err!=nil{return err}};return nil}
func processOne(item string,t *ResourceTracker,handle func(string)error)(err error){r,err:=t.Open();if err!=nil{return err};defer func(){if closeErr:=r.Close();err==nil{err=closeErr}}();return handle(item)}
''')
w('internal/incident/service.go','''
package incident
import("errors";"fmt")
type Service struct{}
func(Service)Record(tx *Tx,validate func()error,write func()error)(err error){if err=validate();err!=nil{_ = tx.Rollback();return fmt.Errorf("validate incident: %w",err)};if err=write();err!=nil{_ = tx.Rollback();return fmt.Errorf("write incident: %w",err)};if err=tx.Commit();err!=nil{_ = tx.Rollback();return fmt.Errorf("commit incident: %w",err)};return nil}
func MergeErrors(primary,cleanup error)error{if primary==nil{return cleanup};if cleanup==nil{return primary};return errors.Join(primary,cleanup)}
''')
w('internal/incident/incident_test.go','''
package incident
import("errors";"fmt";"testing")
func TestIncidentBatchReleasesEachResource(t *testing.T){tracker:=NewResourceTracker(2);items:=make([]string,20);for i:=range items{items[i]=fmt.Sprint(i)};if err:=ProcessBatch(items,tracker,func(string)error{return nil});err!=nil{t.Fatal(err)};if tracker.OpenCount()!=0{t.Fatalf("open=%d",tracker.OpenCount())}}
func TestIncidentBusinessErrorSurvivesCleanup(t *testing.T){business:=errors.New("invalid placard");tx:=NewTx(nil);err:=Service{}.Record(tx,func()error{return business},func()error{return nil});if !errors.Is(err,business){t.Fatalf("business error lost: %v",err)};if !tx.RolledBack()||tx.Committed(){t.Fatalf("tx state committed=%v rolled=%v",tx.Committed(),tx.RolledBack())}}
func TestIncidentCommitFailureRollsBack(t *testing.T){tx:=NewTx(ErrCommit);err:=Service{}.Record(tx,func()error{return nil},func()error{return nil});if !errors.Is(err,ErrCommit){t.Fatalf("commit error lost: %v",err)};if !tx.RolledBack(){t.Fatal("commit failure did not rollback")}}
''')

# BUG 008 dispatch state machine.
w('internal/dispatch/state.go','''
package dispatch
import "fmt"
type State string
const(StateQueued State="queued";StateMoving State="moving";StateHeld State="held";StateRetrying State="retrying";StateCompleted State="completed";StateFailed State="failed")
var transitions=map[State]map[State]bool{StateQueued:{StateMoving:true,StateHeld:true},StateMoving:{StateCompleted:true,StateFailed:true},StateHeld:{StateRetrying:true,StateFailed:true},StateRetrying:{StateMoving:true,StateCompleted:true,StateFailed:true}}
func CanTransition(from,to State)bool{return transitions[from][to]}
func Transition(from,to State)(State,error){if !CanTransition(from,to){return from,fmt.Errorf("invalid dispatch transition %s -> %s",from,to)};return to,nil}
''')
w('internal/dispatch/service.go','''
package dispatch
import "sync"
type Job struct{ ID string; State State; Attempts int }
type Service struct{ mu sync.Mutex; jobs map[string]Job }
func NewService()*Service{return &Service{jobs:make(map[string]Job)}}
func(s *Service)Put(j Job){s.mu.Lock();s.jobs[j.ID]=j;s.mu.Unlock()}
func(s *Service)Move(id string,to State)error{s.mu.Lock();defer s.mu.Unlock();j:=s.jobs[id];next,err:=Transition(j.State,to);if err!=nil{return err};j.State=next;if to==StateRetrying{j.Attempts++};s.jobs[id]=j;return nil}
func(s *Service)Get(id string)(Job,bool){s.mu.Lock();defer s.mu.Unlock();j,ok:=s.jobs[id];return j,ok}
''')
w('internal/dispatch/worker.go','''
package dispatch
type Worker struct{ service *Service }
func NewWorker(s *Service)*Worker{return &Worker{service:s}}
func(w *Worker)Recover(id string)error{if err:=w.service.Move(id,StateRetrying);err!=nil{return err};return w.service.Move(id,StateCompleted)}
''')
w('internal/dispatch/query.go','''
package dispatch
func IsInProgress(s State)bool{return s==StateQueued||s==StateMoving||s==StateHeld||s==StateRetrying}
func FilterInProgress(jobs []Job)[]Job{out:=make([]Job,0,len(jobs));for _,j:=range jobs{if IsInProgress(j.State){out=append(out,j)}};return out}
''')
w('internal/dispatch/dispatch_test.go','''
package dispatch
import "testing"
func TestDispatchRetryCanComplete(t *testing.T){s:=NewService();s.Put(Job{ID:"J",State:StateHeld});if err:=NewWorker(s).Recover("J");err!=nil{t.Fatal(err)};j,_:=s.Get("J");if j.State!=StateCompleted{t.Fatalf("state=%s",j.State)}}
func TestDispatchRetryingIsVisibleInProgress(t *testing.T){got:=FilterInProgress([]Job{{ID:"A",State:StateRetrying},{ID:"B",State:StateCompleted}});if len(got)!=1||got[0].ID!="A"{t.Fatalf("got=%v",got)}}
func TestDispatchRecoveryIncrementsAttemptsOnce(t *testing.T){s:=NewService();s.Put(Job{ID:"J",State:StateHeld});if err:=NewWorker(s).Recover("J");err!=nil{t.Fatal(err)};j,_:=s.Get("J");if j.Attempts!=1{t.Fatalf("attempts=%d",j.Attempts)}}
''')

# BUG 009 yard occupancy snapshot concurrency.
w('internal/occupancy/ledger.go','''
package occupancy
import "sync"
type Slot struct{ Track string; Consist string; HazardClass string; Revision int }
type Ledger struct{ mu sync.RWMutex; slots map[string]Slot; revision int }
func NewLedger()*Ledger{return &Ledger{slots:make(map[string]Slot)}}
func(l *Ledger)Reserve(slot Slot)int{l.mu.Lock();defer l.mu.Unlock();l.revision++;slot.Revision=l.revision;l.slots[slot.Track]=slot;return l.revision}
func(l *Ledger)Release(track string)int{l.mu.Lock();defer l.mu.Unlock();l.revision++;delete(l.slots,track);return l.revision}
func(l *Ledger)Snapshot()(map[string]Slot,int){l.mu.RLock();defer l.mu.RUnlock();out:=make(map[string]Slot,len(l.slots));for k,v:=range l.slots{out[k]=v};return out,l.revision}
''')
w('internal/occupancy/allocator.go','''
package occupancy
import "fmt"
type Allocator struct{ ledger *Ledger }
func NewAllocator(l *Ledger)*Allocator{return &Allocator{ledger:l}}
func(a *Allocator)Move(consist,from,to,class string)error{slots,_:=a.ledger.Snapshot();if current,ok:=slots[from];!ok||current.Consist!=consist{return fmt.Errorf("consist %s not on %s",consist,from)};if _,busy:=slots[to];busy{return fmt.Errorf("track %s occupied",to)};a.ledger.Release(from);a.ledger.Reserve(Slot{Track:to,Consist:consist,HazardClass:class});return nil}
''')
w('internal/occupancy/observer.go','''
package occupancy
type View struct{ Slots map[string]Slot; Revision int; Hazardous int }
func Observe(l *Ledger)View{slots,rev:=l.Snapshot();v:=View{Slots:slots,Revision:rev};for _,s:=range slots{if s.HazardClass!=""{v.Hazardous++}};return v}
func(v View)Clone()View{out:=make(map[string]Slot,len(v.Slots));for k,s:=range v.Slots{out[k]=s};v.Slots=out;return v}
''')
w('internal/occupancy/occupancy_test.go','''
package occupancy
import("fmt";"sync";"testing")
func TestOccupancySnapshotDetached(t *testing.T){l:=NewLedger();l.Reserve(Slot{Track:"T1",Consist:"C1",HazardClass:"3"});s,_:=l.Snapshot();delete(s,"T1");again,_:=l.Snapshot();if _,ok:=again["T1"];!ok{t.Fatal("external map mutation changed ledger")}}
func TestOccupancyConcurrentViewsAreConsistent(t *testing.T){l:=NewLedger();start:=make(chan struct{});var wg sync.WaitGroup;wg.Add(2);go func(){defer wg.Done();<-start;for i:=0;i<300;i++{track:=fmt.Sprintf("T%d",i%8);l.Reserve(Slot{Track:track,Consist:fmt.Sprint(i),HazardClass:"3"})}}();go func(){defer wg.Done();<-start;for i:=0;i<300;i++{v:=Observe(l);if v.Hazardous!=len(v.Slots){t.Errorf("torn view hazardous=%d slots=%d",v.Hazardous,len(v.Slots));return}}}();close(start);wg.Wait()}
func TestOccupancyRevisionCoversPublishedSlots(t *testing.T){l:=NewLedger();for i:=0;i<20;i++{l.Reserve(Slot{Track:fmt.Sprint(i),Consist:fmt.Sprint(i)})};v:=Observe(l);if v.Revision!=len(v.Slots){t.Fatalf("revision=%d slots=%d",v.Revision,len(v.Slots))}}
''')

# BUG 010 audit stream cancellation/checkpoint.
w('internal/audit/checkpoint.go','''
package audit
import "sync"
type Checkpoint struct{ mu sync.Mutex; value int }
func(c *Checkpoint)Load()int{c.mu.Lock();defer c.mu.Unlock();return c.value}
func(c *Checkpoint)Commit(v int){c.mu.Lock();if v>c.value{c.value=v};c.mu.Unlock()}
''')
w('internal/audit/subscriber.go','''
package audit
import "context"
type Event struct{ Sequence int; Kind string }
type Source interface{ Next(context.Context,int)(Event,error) }
type SourceFunc func(context.Context,int)(Event,error)
func(f SourceFunc)Next(c context.Context,n int)(Event,error){return f(c,n)}
type Sink interface{ Deliver(context.Context,Event)error }
type SinkFunc func(context.Context,Event)error
func(f SinkFunc)Deliver(c context.Context,e Event)error{return f(c,e)}
''')
w('internal/audit/stream.go','''
package audit
import("context";"fmt")
type Stream struct{ source Source; sink Sink; checkpoint *Checkpoint }
func NewStream(src Source,sink Sink,cp *Checkpoint)*Stream{return &Stream{source:src,sink:sink,checkpoint:cp}}
func(s *Stream)Run(ctx context.Context,limit int)error{seq:=s.checkpoint.Load();for n:=0;n<limit;n++{if err:=ctx.Err();err!=nil{return err};event,err:=s.source.Next(ctx,seq+1);if err!=nil{return fmt.Errorf("read audit event %d: %w",seq+1,err)};if err=s.sink.Deliver(ctx,event);err!=nil{return fmt.Errorf("deliver audit event %d: %w",event.Sequence,err)};s.checkpoint.Commit(event.Sequence);seq=event.Sequence};return nil}
''')
w('internal/audit/audit_test.go','''
package audit
import("context";"errors";"sync/atomic";"testing";"time")
func TestAuditCancellationReachesSource(t *testing.T){seen:=make(chan struct{});src:=SourceFunc(func(ctx context.Context,n int)(Event,error){<-ctx.Done();close(seen);return Event{},ctx.Err()});s:=NewStream(src,SinkFunc(func(context.Context,Event)error{return nil}),&Checkpoint{});ctx,cancel:=context.WithTimeout(context.Background(),10*time.Millisecond);defer cancel();err:=s.Run(ctx,1);if !errors.Is(err,context.DeadlineExceeded){t.Fatalf("err=%v",err)};select{case <-seen:case <-time.After(time.Second):t.Fatal("source missed cancellation")}}
func TestAuditFailedDeliveryDoesNotAdvanceCheckpoint(t *testing.T){cp:=&Checkpoint{};src:=SourceFunc(func(context.Context,int)(Event,error){return Event{Sequence:1,Kind:"move"},nil});s:=NewStream(src,SinkFunc(func(context.Context,Event)error{return errors.New("down")}),cp);_ = s.Run(context.Background(),1);if cp.Load()!=0{t.Fatalf("checkpoint=%d",cp.Load())}}
func TestAuditStopsReadingAfterCancel(t *testing.T){var reads atomic.Int32;ctx,cancel:=context.WithCancel(context.Background());src:=SourceFunc(func(context.Context,int)(Event,error){n:=reads.Add(1);if n==1{cancel()};return Event{Sequence:int(n)},nil});s:=NewStream(src,SinkFunc(func(context.Context,Event)error{return nil}),&Checkpoint{});_ = s.Run(ctx,10);if reads.Load()!=1{t.Fatalf("reads=%d",reads.Load())}}
''')

# Generate substantial vertical-industry policy code (functional, non-test).
policies = [
('placard','Placard','PlacardCode'),('separation','Separation','HazardClass'),('brake','Brake','BrakeRatio'),('route','Route','RouteCode'),('crew','Crew','Certification'),('weather','Weather','WindKPH'),('temperature','Temperature','Celsius'),('seal','Seal','SealCode'),('weight','Weight','GrossTons'),('length','Length','Meters'),('track','Track','TrackCode'),('lighting','Lighting','Lux'),('radio','Radio','Channel'),('escort','Escort','EscortID'),('document','Document','DocumentID'),('speed','Speed','KPH'),('gradient','Gradient','Permille'),('buffer','Buffer','Meters'),('alarm','Alarm','AlarmCode'),('maintenance','Maintenance','DueHours'),('handover','Handover','OperatorID'),('sampling','Sampling','SampleID'),('containment','Containment','Capacity'),('ventilation','Ventilation','AirChanges'),('grounding','Grounding','Resistance'),('firewater','Firewater','Pressure'),('spillkit','SpillKit','Units'),('access','Access','Zone'),('crossing','Crossing','CrossingID'),('notification','Notification','Contact'),('documentation','Documentation','PacketID'),('locomotive','Locomotive','UnitID'),('clearancewindow','ClearanceWindow','Minutes'),('sidings','Siding','SidingID'),('detector','Detector','DetectorID'),('weighbridge','Weighbridge','TicketID'),('switchlock','SwitchLock','LockID'),('emergency','Emergency','PlanID'),('training','Training','CourseID'),('insurance','Insurance','PolicyID')]
for idx,(pkg,typ,field) in enumerate(policies,1):
    numeric = field in {'BrakeRatio','WindKPH','Celsius','GrossTons','Meters','Lux','KPH','Permille','DueHours','Capacity','AirChanges','Resistance','Pressure','Units','Minutes'}
    ft='float64' if numeric else 'string'
    zero='0' if numeric else '""'
    cond=f'r.{field} <= 0' if numeric else f'strings.TrimSpace(r.{field}) == ""'
    imports='"fmt"\n    "sort"\n    "strings"'
    w(f'internal/policy/{pkg}.go',f'''
package policy

import (
    {imports}
)

type {typ}Record struct {{
    Yard string
    ConsistID string
    {field} {ft}
    Active bool
    Tags []string
    Revision int
}}

type {typ}Decision struct {{
    Allowed bool
    Reasons []string
    NormalizedTags []string
    Score int
}}

func Evaluate{typ}(r {typ}Record) {typ}Decision {{
    reasons := make([]string, 0, 4)
    if strings.TrimSpace(r.Yard) == "" {{ reasons = append(reasons, "yard is required") }}
    if strings.TrimSpace(r.ConsistID) == "" {{ reasons = append(reasons, "consist is required") }}
    if {cond} {{ reasons = append(reasons, "{field} is required") }}
    if !r.Active {{ reasons = append(reasons, "record is inactive") }}
    tags := normalize{typ}Tags(r.Tags)
    score := 100 - len(reasons)*20 + min{typ}(len(tags)*2, 10)
    if score < 0 {{ score = 0 }}
    return {typ}Decision{{Allowed: len(reasons)==0, Reasons: reasons, NormalizedTags: tags, Score: score}}
}}

func Validate{typ}Batch(records []{typ}Record) error {{
    seen := make(map[string]struct{{}}, len(records))
    for i, r := range records {{
        key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.{field})
        if _, ok := seen[key]; ok {{ return fmt.Errorf("duplicate {pkg} record at index %d", i) }}
        seen[key] = struct{{}}{{}}
        if d := Evaluate{typ}(r); !d.Allowed {{ return fmt.Errorf("invalid {pkg} record %d: %s", i, strings.Join(d.Reasons, ", ")) }}
    }}
    return nil
}}

func normalize{typ}Tags(tags []string) []string {{
    unique := make(map[string]struct{{}}, len(tags))
    for _, tag := range tags {{ tag = strings.ToLower(strings.TrimSpace(tag)); if tag != "" {{ unique[tag]=struct{{}}{{}} }} }}
    out := make([]string,0,len(unique)); for tag := range unique {{ out=append(out,tag) }}; sort.Strings(out); return out
}}
func min{typ}(a,b int)int{{if a<b{{return a}};return b}}
''')

# Additional operational packages to deepen project structure.
ops=['gate','switching','classification','marshalling','weighing','samplingops','firewatch','security','handoff','compliance','forecast','capacity']
for i,pkg in enumerate(ops,1):
    typ=''.join(x.title() for x in pkg.split('_'))
    w(f'internal/{pkg}/service.go',f'''
package {pkg}
import("fmt";"sort";"strings";"sync";"time")
type Record struct{{ID string;Yard string;Status string;Priority int;UpdatedAt time.Time;Labels []string}}
type Service struct{{mu sync.RWMutex;records map[string]Record;order []string}}
func NewService()*Service{{return &Service{{records:make(map[string]Record)}}}}
func(s *Service)Upsert(r Record)error{{r.ID=strings.TrimSpace(r.ID);r.Yard=strings.TrimSpace(r.Yard);if r.ID==""||r.Yard==""{{return fmt.Errorf("{pkg}: id and yard required")}};if r.UpdatedAt.IsZero(){{r.UpdatedAt=time.Now().UTC()}};r.Labels=normalize(r.Labels);s.mu.Lock();defer s.mu.Unlock();if _,ok:=s.records[r.ID];!ok{{s.order=append(s.order,r.ID)}};s.records[r.ID]=clone(r);return nil}}
func(s *Service)Get(id string)(Record,bool){{s.mu.RLock();defer s.mu.RUnlock();r,ok:=s.records[id];return clone(r),ok}}
func(s *Service)List(yard,status string)[]Record{{s.mu.RLock();defer s.mu.RUnlock();out:=make([]Record,0,len(s.records));for _,id:=range s.order{{r:=s.records[id];if yard!=""&&r.Yard!=yard{{continue}};if status!=""&&r.Status!=status{{continue}};out=append(out,clone(r))}};sort.SliceStable(out,func(i,j int)bool{{if out[i].Priority==out[j].Priority{{return out[i].UpdatedAt.Before(out[j].UpdatedAt)}};return out[i].Priority>out[j].Priority}});return out}}
func(s *Service)Transition(id,from,to string)error{{s.mu.Lock();defer s.mu.Unlock();r,ok:=s.records[id];if !ok{{return fmt.Errorf("{pkg}: record %s missing",id)}};if from!=""&&r.Status!=from{{return fmt.Errorf("{pkg}: expected %s got %s",from,r.Status)}};if strings.TrimSpace(to)==""{{return fmt.Errorf("{pkg}: empty target status")}};r.Status=to;r.UpdatedAt=time.Now().UTC();s.records[id]=r;return nil}}
func clone(r Record)Record{{r.Labels=append([]string(nil),r.Labels...);return r}}
func normalize(in []string)[]string{{m:=map[string]struct{{}}{{}};for _,v:=range in{{v=strings.ToLower(strings.TrimSpace(v));if v!=""{{m[v]=struct{{}}{{}}}}}};out:=make([]string,0,len(m));for v:=range m{{out=append(out,v)}};sort.Strings(out);return out}}
''')

print(root)
