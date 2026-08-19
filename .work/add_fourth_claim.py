from pathlib import Path
import shutil, textwrap
root=Path('/Users/gaobo/repositories/gitlab/评审项目/bug-go项目标注/0819')
source=root/'2026-08-19/rail-hazmat-yard-control-service'

def w(base,rel,s):
 p=base/rel;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(textwrap.dedent(s).lstrip())
def append(base,rel,s):
 p=base/rel;txt=p.read_text();marker=s.strip().splitlines()[0]
 if marker not in txt:p.write_text(txt.rstrip()+"\n\n"+textwrap.dedent(s).lstrip())

goods={
1:('internal/manifest/publisher.go','''
package manifest
import "sync"
type Publisher struct{mu sync.RWMutex;view View}
func(p *Publisher)Publish(v View){p.mu.Lock();p.view=v.Clone();p.mu.Unlock()}
func(p *Publisher)Snapshot()View{p.mu.RLock();defer p.mu.RUnlock();return p.view.Clone()}
''','internal/manifest/manifest_test.go','''
func TestManifestPublisherDetachesView(t *testing.T){p:=&Publisher{};v:=NewView([]Car{{ID:"A",UNNumber:"1203",Seal:"S1"}},7);p.Publish(v);v.Cars[0].Seal="input";first:=p.Snapshot();first.Cars[0].Seal="output";again:=p.Snapshot();if again.Cars[0].Seal!="S1"{t.Fatalf("publisher view polluted: %s",again.Cars[0].Seal)}}
''','''
package manifest
import "sync"
type Publisher struct{mu sync.RWMutex;view View}
func(p *Publisher)Publish(v View){p.mu.Lock();p.view=v;p.mu.Unlock()}
func(p *Publisher)Snapshot()View{p.mu.RLock();defer p.mu.RUnlock();return p.view}
'''),
2:('internal/inspection/report.go','''
package inspection
import "sync"
type ReportStore struct{mu sync.RWMutex;results []Result;errs []error}
func(s *ReportStore)Publish(results []Result,errs []error){s.mu.Lock();s.results=append([]Result(nil),results...);s.errs=append([]error(nil),errs...);s.mu.Unlock()}
func(s *ReportStore)Snapshot()([]Result,[]error){s.mu.RLock();defer s.mu.RUnlock();return append([]Result(nil),s.results...),append([]error(nil),s.errs...)}
''','internal/inspection/inspection_test.go','''
func TestInspectionReportPublicationDetached(t *testing.T){store:=&ReportStore{};results:=[]Result{{CarID:"A",Passed:true}};errs:=[]error{errors.New("gate warning")};store.Publish(results,errs);results[0].CarID="input";errs[0]=nil;got,gotErrs:=store.Snapshot();got[0].CarID="output";gotErrs[0]=nil;again,againErrs:=store.Snapshot();if again[0].CarID!="A"||againErrs[0]==nil{t.Fatalf("published report was mutated: results=%v errors=%v",again,againErrs)}}
''','''
package inspection
import "sync"
type ReportStore struct{mu sync.RWMutex;results []Result;errs []error}
func(s *ReportStore)Publish(results []Result,errs []error){s.mu.Lock();s.results=results;s.errs=errs;s.mu.Unlock()}
func(s *ReportStore)Snapshot()([]Result,[]error){s.mu.RLock();defer s.mu.RUnlock();return s.results,s.errs}
'''),
3:('internal/permit/classifier.go','''
package permit
import("errors";"rail-hazmat-yard-control-service/internal/platform")
func ShouldRetry(err error)bool{return errors.Is(err,ErrAuthorityBusy)||errors.Is(err,platform.ErrUnavailable)}
''','internal/permit/permit_test.go','''
func TestPermitClassifierHandlesWrappedBusy(t *testing.T){wrapped:=fmt.Errorf("authority response: %w",ErrAuthorityBusy);if !ShouldRetry(wrapped){t.Fatal("wrapped busy error classified as terminal")};if ShouldRetry(fmt.Errorf("denied: %w",ErrPermitDenied)){t.Fatal("denial classified as retryable")}}
''','''
package permit
import "rail-hazmat-yard-control-service/internal/platform"
func ShouldRetry(err error)bool{if err==nil{return false};return err.Error()==ErrAuthorityBusy.Error()||err.Error()==platform.ErrUnavailable.Error()}
'''),
4:('internal/telemetry/probe.go','''
package telemetry
func DecoderAvailable(d Decoder)bool{return !isNilDecoder(d)}
''','internal/telemetry/telemetry_test.go','''
func TestTelemetryProbeRejectsTypedNil(t *testing.T){var d *nilDecoder;if DecoderAvailable(d){t.Fatal("probe accepted typed nil decoder")}}
''','''
package telemetry
func DecoderAvailable(d Decoder)bool{return d!=nil}
'''),
5:('internal/clearance/session.go','''
package clearance
import "context"
func RunSession(ctx context.Context,p *Pipeline,id string)error{if err:=ctx.Err();err!=nil{return err};return p.Clear(ctx,id)}
''','internal/clearance/clearance_test.go','''
func TestClearanceSessionUsesCallerContext(t *testing.T){var calls atomic.Int32;p:=NewPipeline(NewRepository(),VerifierFunc(func(context.Context,string)error{calls.Add(1);return nil}));ctx,cancel:=context.WithCancel(context.Background());cancel();err:=RunSession(ctx,p,"C9");if !errors.Is(err,context.Canceled){t.Fatalf("err=%v",err)};if calls.Load()!=0{t.Fatalf("verifier called %d times",calls.Load())}}
''','''
package clearance
import "context"
func RunSession(ctx context.Context,p *Pipeline,id string)error{return p.Clear(context.Background(),id)}
'''),
6:('internal/consist/export.go','''
package consist
func ExportPlan(p Plan)Plan{return p.Clone()}
''','internal/consist/consist_test.go','''
func TestConsistExportDoesNotExposePlanStorage(t *testing.T){plan:=BuildPlan([]Wagon{{"A",true,""}},"T4");exported:=ExportPlan(plan);exported.Routed[0].Track="BAD";if plan.Routed[0].Track!="T4"{t.Fatalf("export mutation leaked into plan: %#v",plan)}}
''','''
package consist
func ExportPlan(p Plan)Plan{return p}
'''),
7:('internal/incident/cleanup.go','''
package incident
func Finalize(primary error,closeFn func()error)error{if closeFn==nil{return primary};return MergeErrors(primary,closeFn())}
''','internal/incident/incident_test.go','''
func TestIncidentCleanupKeepsPrimaryError(t *testing.T){primary:=errors.New("placard mismatch");cleanup:=errors.New("close failed");err:=Finalize(primary,func()error{return cleanup});if !errors.Is(err,primary)||!errors.Is(err,cleanup){t.Fatalf("combined error lost component: %v",err)}}
''','''
package incident
func Finalize(primary error,closeFn func()error)error{if closeFn==nil{return primary};if err:=closeFn();err!=nil{return err};return primary}
'''),
8:('internal/dispatch/projection.go','''
package dispatch
type Projection struct{State State;InProgress bool;Attempts int}
func Project(j Job)Projection{return Projection{State:j.State,InProgress:IsInProgress(j.State),Attempts:j.Attempts}}
''','internal/dispatch/dispatch_test.go','''
func TestDispatchProjectionPreservesRetrying(t *testing.T){p:=Project(Job{ID:"J",State:StateRetrying,Attempts:2});if p.State!=StateRetrying||!p.InProgress||p.Attempts!=2{t.Fatalf("projection=%+v",p)}}
''','''
package dispatch
type Projection struct{State State;InProgress bool;Attempts int}
func Project(j Job)Projection{state:=j.State;if state==StateRetrying{state=StateMoving};return Projection{State:state,InProgress:IsInProgress(state),Attempts:j.Attempts+1}}
'''),
9:('internal/occupancy/cache.go','''
package occupancy
import "sync"
type Cache struct{mu sync.RWMutex;view View}
func(c *Cache)Publish(v View){c.mu.Lock();c.view=v.Clone();c.mu.Unlock()}
func(c *Cache)Snapshot()View{c.mu.RLock();defer c.mu.RUnlock();return c.view.Clone()}
''','internal/occupancy/occupancy_test.go','''
func TestOccupancyCacheDoesNotExposePublishedMap(t *testing.T){c:=&Cache{};v:=View{Slots:map[string]Slot{"T1":{Track:"T1",Consist:"C1"}},Revision:3};c.Publish(v);delete(v.Slots,"T1");first:=c.Snapshot();delete(first.Slots,"T1");again:=c.Snapshot();if _,ok:=again.Slots["T1"];!ok{t.Fatal("cache view map was externally mutated")}}
''','''
package occupancy
import "sync"
type Cache struct{mu sync.RWMutex;view View}
func(c *Cache)Publish(v View){c.mu.Lock();c.view=v;c.mu.Unlock()}
func(c *Cache)Snapshot()View{c.mu.RLock();defer c.mu.RUnlock();return c.view}
'''),
10:('internal/audit/delivery.go','''
package audit
import "context"
func DeliverAndCommit(ctx context.Context,sink Sink,cp *Checkpoint,event Event)error{if err:=ctx.Err();err!=nil{return err};if err:=sink.Deliver(ctx,event);err!=nil{return err};cp.Commit(event.Sequence);return nil}
''','internal/audit/audit_test.go','''
func TestAuditDeliveryHelperCommitsAfterSuccess(t *testing.T){cp:=&Checkpoint{};err:=DeliverAndCommit(context.Background(),SinkFunc(func(context.Context,Event)error{return errors.New("offline")}),cp,Event{Sequence:9});if err==nil{t.Fatal("expected delivery error")};if cp.Load()!=0{t.Fatalf("checkpoint advanced to %d",cp.Load())}}
''','''
package audit
import "context"
func DeliverAndCommit(ctx context.Context,sink Sink,cp *Checkpoint,event Event)error{cp.Commit(event.Sequence);return sink.Deliver(context.Background(),event)}
''')}

for n,(file,good,testfile,test,bad) in goods.items():
 w(source,file,good);append(source,testfile,test)
 for k in range(1,11):
  rec=f'rail-hazmat-yard-control-service__{k:03d}'
  w(root/'_gold'/rec,file,good);append(root/'_gold'/rec,testfile,test)
  env=root/'2026-08-19'/rec/'env';w(env,file,bad if k==n else good);append(env,testfile,test)
print('done')
