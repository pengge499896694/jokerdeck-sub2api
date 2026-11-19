"""Read-only final verification; publish a redacted release report."""
import datetime, hashlib, json, pathlib, runpy, subprocess
s=runpy.run_path('/opt/sub2api/staging/v0.2.23-20261119/deploy.py',run_name='postflight')
stage=s['STAGE']; new=s['inspect'](s['NEW']);old=s['inspect'](s['OLD']);edge=s['inspect']('edge-nginx')
observation=json.loads((stage/'observation.json').read_text())
assert observation['seconds']>=600 and observation['checks']>0
assert observation['nginx_statuses'].get('502',0)==observation['nginx_statuses'].get('504',0)==0
assert observation['fatal_log_lines']==0 and observation['nginx_upstream_errors']==0
assert edge['State']['StartedAt']=='2026-10-27T15:53:31.118275501Z','Nginx was restarted'
assert edge['RestartCount']==0
s['health'](s['NEW'])
assert new['State']['Health']['Status']=='healthy' and new['RestartCount']==0
assert s['CONF'].read_bytes()==(s['BACKUP']/'nginx.conf').read_bytes().replace(s['OLD_LINE'],s['NEW_LINE'])
expected=(stage/'SHA256SUMS').read_text().split()[0]
assert s['out'](['docker','exec',s['NEW'],'sha256sum','/app/sub2api']).split()[0]==expected
before=json.loads((s['BACKUP']/'active-inspect.json').read_text())
assert dict(x.split('=',1) for x in before['Config']['Env'])==dict(x.split('=',1) for x in new['Config']['Env'])
mounts=lambda x:sorted(tuple(m[k] for k in ('Type','Source','Destination','RW','Propagation')) for m in x['Mounts'])
assert mounts(before)==mounts(new),'Mount configuration drift'
assert set(before['NetworkSettings']['Networks'])==set(new['NetworkSettings']['Networks'])
assert not new['NetworkSettings']['Ports'].get('8080/tcp'),'Candidate should not publish an unprotected host port'
baseline=json.loads((stage/'all-sites-baseline.json').read_text());origins={}
for host in baseline:
 origins[host]=s['curl_status'](host,origin=True)
 assert origins[host]==baseline[host]==200,host
public={host:s['curl_status'](host) for host in s['HOSTS']}
assert all(v==200 for v in public.values())
probes={path:s['curl_status']('api.jokere.asia',path,data='{}') for path in ('/v1/messages','/v1/chat/completions','/v1/responses')}
assert all(v==401 for v in probes.values())
rows=s['out'](['docker','exec','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-At','-F','|','-c','SELECT filename,checksum FROM schema_migrations'])
assert dict(l.split('|') for l in rows.splitlines())==json.loads((stage/'migrations.json').read_text())
statuses,gate,transport,fatal=s['log_report']()
assert statuses['502']==statuses['504']==0 and not transport and not fatal
preworkers=set(json.loads((stage/'pre-switch-workers.json').read_text()));remaining=preworkers&s['workers']()
assert not remaining or old['State']['Running'],'Previous app must stay running if old workers exist'
subprocess.run(['docker','compose','--env-file',str(s['ROOT']/'current.env'),'-f',str(s['ROOT']/'docker-compose.yml'),'config','--quiet'],check=True)
report={
 'verified_at':datetime.datetime.now(datetime.timezone.utc).isoformat(),
 'cutover_at':(stage/'switch-time.txt').read_text().strip(),
 'version':'0.2.23','code_commit':json.loads((stage/'build-info.json').read_text())['commit'],
 'binary_sha256':expected,'active_container':s['NEW'],'image':new['Config']['Image'],
 'image_version_label':new['Config']['Labels']['org.opencontainers.image.version'],
 'old_container_running':old['State']['Running'],'old_retained':True,'old_nginx_workers_remaining':len(remaining),
 'nginx_original_started_at':edge['State']['StartedAt'],'nginx_not_restarted':True,
 'candidate_restarts':new['RestartCount'],'candidate_health':new['State']['Health']['Status'],
 'all_origin_sites':origins,'representative_public_sites':public,'gateway_unauthenticated_probes':probes,
 'api_checks':len(json.loads((stage/'verification.json').read_text())),
 'old_asset_hashes_preserved':len(json.loads((stage/'old-assets.json').read_text())),
 'migration_count':len(rows.splitlines()),'environment_preserved':True,'mounts_preserved':True,'networks_preserved':True,
 'nginx_statuses_since_cutover':dict(statuses),'real_gateway_statuses_since_cutover':dict(gate),
 'nginx_upstream_transport_errors':transport,'candidate_error_or_fatal_lines':len(fatal),
 'backup_dir':str(s['BACKUP']),'rollback_command':'python3 '+str(stage/'rollback.py'),
 'paid_synthetic_gateway_requests':0,
}
(stage/'final-report.json').write_text(json.dumps(report,indent=2));print(json.dumps(report,ensure_ascii=False),flush=True)
