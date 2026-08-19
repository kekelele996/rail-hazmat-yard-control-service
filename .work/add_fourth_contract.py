from pathlib import Path
import json
base=Path('/Users/gaobo/repositories/gitlab/评审项目/bug-go项目标注/0819/2026-08-19')
extra={
1:("go test -race ./internal/manifest -run '^TestManifestPublisherDetachesView$' -count=1",'published_view_detached','发布器输入输出都不能暴露可变清单','TestManifestPublisherDetachesView',['internal/manifest/publisher.go']),
2:("go test -race ./internal/inspection -run '^TestInspectionReportPublicationDetached$' -count=1",'report_publication_detached','检查报告发布后与调用方切片隔离','TestInspectionReportPublicationDetached',['internal/inspection/report.go']),
3:("go test ./internal/permit -run '^TestPermitClassifierHandlesWrappedBusy$' -count=1",'wrapped_busy_classified','被包装的繁忙错误仍可重试且拒绝不可重试','TestPermitClassifierHandlesWrappedBusy',['internal/permit/classifier.go']),
4:("go test ./internal/telemetry -run '^TestTelemetryProbeRejectsTypedNil$' -count=1",'probe_rejects_typed_nil','可用性探测必须识别 typed nil decoder','TestTelemetryProbeRejectsTypedNil',['internal/telemetry/probe.go','internal/telemetry/validator.go']),
5:("go test ./internal/clearance -run '^TestClearanceSessionUsesCallerContext$' -count=1",'session_uses_caller_context','会话入口不能用后台 context 替换调用方取消','TestClearanceSessionUsesCallerContext',['internal/clearance/session.go','internal/clearance/pipeline.go']),
6:("go test ./internal/consist -run '^TestConsistExportDoesNotExposePlanStorage$' -count=1",'export_plan_detached','导出计划不得暴露内部切片存储','TestConsistExportDoesNotExposePlanStorage',['internal/consist/export.go','internal/consist/planner.go']),
7:("go test ./internal/incident -run '^TestIncidentCleanupKeepsPrimaryError$' -count=1",'cleanup_error_merged','清理失败与主错误都要保留','TestIncidentCleanupKeepsPrimaryError',['internal/incident/cleanup.go','internal/incident/service.go']),
8:("go test ./internal/dispatch -run '^TestDispatchProjectionPreservesRetrying$' -count=1",'projection_preserves_retrying','状态投影不能改写 retrying 或尝试次数','TestDispatchProjectionPreservesRetrying',['internal/dispatch/projection.go','internal/dispatch/query.go']),
9:("go test -race ./internal/occupancy -run '^TestOccupancyCacheDoesNotExposePublishedMap$' -count=1",'cache_map_detached','占用缓存发布和读取均隔离 map','TestOccupancyCacheDoesNotExposePublishedMap',['internal/occupancy/cache.go','internal/occupancy/observer.go']),
10:("go test ./internal/audit -run '^TestAuditDeliveryHelperCommitsAfterSuccess$' -count=1",'delivery_before_checkpoint','交付成功以后才能提交位点','TestAuditDeliveryHelperCommitsAfterSuccess',['internal/audit/delivery.go','internal/audit/checkpoint.go'])}
for n,item in extra.items():
 rec=base/f'rail-hazmat-yard-control-service__{n:03d}'
 for filename in ('collection.json','draft_collection.json'):
  p=rec/filename
  if not p.exists():continue
  d=json.loads(p.read_text());cmd,cid,desc,test,paths=item
  cmds=[x for x in str(d['verify_cmds']).splitlines() if x.strip()]
  if cmd not in cmds:cmds.append(cmd)
  d['verify_cmds']='\n'.join(cmds)
  claims=d['coverage_contract']['claims']
  if not any(c['id']==cid for c in claims):claims.append({'id':cid,'description':desc,'test':test,'command':cmd,'paths':paths})
  p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
