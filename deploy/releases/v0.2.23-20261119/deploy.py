#!/usr/bin/env python3
"""Guarded blue/green release for the existing containerized edge topology."""
import base64,collections,datetime,gzip,hashlib,hmac,json,os,pathlib,re,secrets,subprocess,sys,time,urllib.request,urllib.error,urllib.parse
os.umask(0o077)
ROOT=pathlib.Path('/opt/sub2api'); STAGE=ROOT/'staging/v0.2.23-20261119'; BACKUP=ROOT/'backups/pre-v0.2.23-20261119'
OLD='sub2api-v0.2.21-20261112'; NEW='sub2api-v0.2.23-20261119'; IMAGE='local/sub2api:v0.2.23-jokerdeck-20261119'
CONF=pathlib.Path('/opt/edge/nginx.conf'); HOSTS=('jokerdeck.de5.net','jokerdeck.cc.cd','api.jokere.asia'); UPSTREAMS=11
OLD_LINE=f'set $sub2api_upstream http://{OLD}:8080;'.encode(); NEW_LINE=f'set $sub2api_upstream http://{NEW}:8080;'.encode()
CLIENT=urllib.request.build_opener(urllib.request.ProxyHandler({}))
def run(a,**kw):return subprocess.run(a,check=True,**kw)
def out(a,**kw):return subprocess.check_output(a,text=True,**kw)
def inspect(n):return json.loads(out(['docker','inspect',n]))[0]
def base(n):return 'http://'+inspect(n)['NetworkSettings']['Networks']['public-proxy']['IPAddress']+':8080'
def req(url,token=None,data=None):
 headers={'User-Agent':'Jokerdeck-Upgrade-Check/0.2.23'}
 if token:headers['Authorization']='Bearer '+token
 if data is not None:headers['Content-Type']='application/json'
 try:
  with CLIENT.open(urllib.request.Request(url,headers=headers,data=data),timeout=30) as r:return r.status,r.read(),dict(r.headers)
 except urllib.error.HTTPError as e:return e.code,e.read(),dict(e.headers)
def health(n):
 x=inspect(n);assert x['State']['Running'] and x['RestartCount']==0,(n,'not stable')
 run(['docker','exec','edge-nginx','wget','-q','-T','5','-O','/dev/null',f'http://{n}:8080/health'])
def workers():
 result=set()
 for line in out(['docker','exec','edge-nginx','ps','-o','pid,args']).splitlines():
  if 'nginx: worker process' in line:result.add(int(line.split()[0]))
 return result

def edge_reload(content):
 inode=CONF.stat().st_ino
 with CONF.open('wb') as f:f.write(content);f.flush();os.fsync(f.fileno())
 assert CONF.stat().st_ino==inode,'bind mount inode changed'
 assert subprocess.check_output(['docker','exec','edge-nginx','cat','/etc/nginx/conf.d/default.conf'])==content
 run(['docker','exec','edge-nginx','nginx','-t'])
 run(['docker','exec','edge-nginx','nginx','-s','reload'])
 actual=out(['docker','exec','edge-nginx','nginx','-T'],stderr=subprocess.DEVNULL).encode();assert content in actual

def curl_status(host,path='/health',origin=False,data=None):
 a=['curl','-sS','-A','Jokerdeck-Upgrade-Check/0.2.23','--connect-timeout','3','--max-time','10','-o','/dev/null','-w','%{http_code}']
 if origin:a+=['--resolve',host+':443:127.0.0.1']
 if data is not None:a+=['-H','Content-Type: application/json','--data-binary',data]
 a += [f'https://{host}{path}']
 for attempt in range(3):
  try:return int(out(a))
  except subprocess.CalledProcessError as e:
   # A transient DNS/connection timeout is not an HTTP failure. Retry it;
   # HTTP statuses are returned immediately and remain subject to observation.
   if e.returncode not in (6,28) or attempt==2:raise
   time.sleep(1)

def token():
 row=out(['docker','exec','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-Atc',"SELECT id,email,role,password_hash FROM users WHERE role IN ('admin','super_admin') AND status='active' ORDER BY id LIMIT 1"]).strip().split('|')
 assert len(row)==4
 env=dict(x.split('=',1) for x in inspect(OLD)['Config']['Env'] if '=' in x)
 fingerprint=int.from_bytes(hashlib.sha256((row[1].strip().lower()+'\n'+row[3]).encode()).digest()[:8],'big') & 0x7fffffffffffffff
 b=lambda x:base64.urlsafe_b64encode(x).rstrip(b'=').decode();now=int(time.time())
 p={'user_id':int(row[0]),'email':row[1],'role':row[2],'token_version':fingerprint,'sid':secrets.token_hex(8),'iat':now,'nbf':now,'exp':now+600}
 body=(b(b'{"alg":"HS256","typ":"JWT"}')+'.'+b(json.dumps(p,separators=(',',':')).encode())).encode()
 return body.decode()+'.'+b(hmac.new(env['JWT_SECRET'].encode(),body,hashlib.sha256).digest())

def start():
 assert json.loads((STAGE/'rehearsal.json').read_text())['full_backup_restored']
 x=json.loads((BACKUP/'active-inspect.json').read_text());health(OLD)
 expected=(STAGE/'SHA256SUMS').read_text().split()[0]
 with gzip.open(STAGE/'sub2api.gz','rb') as f:data=f.read()
 assert hashlib.sha256(data).hexdigest()==expected
 (STAGE/'sub2api').write_bytes(data);(STAGE/'sub2api').chmod(0o755)
 assert not out(['docker','ps','-a','--filter',f'name=^/{NEW}$','--format','{{.Names}}']).strip()
 # Build before touching the production schema. Pin the existing image digest.
 resource_manifest=json.loads((STAGE/'resources-manifest.json').read_text())
 assert resource_manifest and all(path.startswith('resources/') and '..' not in pathlib.PurePosixPath(path).parts for path in resource_manifest)
 for path,digest in resource_manifest.items():
  assert hashlib.sha256((STAGE/path).read_bytes()).hexdigest()==digest,('resource checksum mismatch',path)
  (STAGE/path).chmod(0o644)
 base_image='local/sub2api:base-v0.2.23-20261119'
 run(['docker','tag',x['Image'],base_image])
 assert out(['docker','image','inspect',base_image,'--format','{{.Id}}']).strip()==x['Image'],'base image digest drift'
 build=json.loads((STAGE/'build-info.json').read_text());assert build['version']=='0.2.23'
 (STAGE/'Dockerfile').write_text('FROM '+base_image+'\nUSER root\nCOPY --chown=1000:1000 sub2api /app/sub2api\nCOPY --chown=1000:1000 resources /app/resources\nLABEL org.opencontainers.image.version="0.2.23" org.opencontainers.image.revision="'+build['commit']+'"\nRUN chmod 0755 /app/sub2api\n')
 run(['docker','build','--pull=false','-t',IMAGE,str(STAGE)])
 # Apply only the two verified additive migrations, with bounded lock waits.
 current=dict(line.split('|') for line in out(['docker','exec','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-At','-F','|','-c','SELECT filename,checksum FROM schema_migrations']).splitlines())
 previous=json.loads((BACKUP/'migrations.json').read_text()); migration_manifest=json.loads((STAGE/'migrations.json').read_text())
 names=['251_group_force_openai_ultrafast.sql','251_group_usage_rollup_nonblocking_build.sql','252_channel_pricing_ultrafast_multiplier.sql']
 assert all(migration_manifest.get(k)==v for k,v in previous.items()) and set(migration_manifest)-set(previous)==set(names)
 assert all(current.get(k)==v for k,v in previous.items()) and set(current)-set(previous)<=set(names)
 for name in names:
  sql=(STAGE/name).read_text().strip();checksum=hashlib.sha256(sql.encode()).hexdigest();assert migration_manifest[name]==checksum
  if name in current:assert current[name]==checksum;continue
  if name.endswith('_notx.sql'):
   # psql reads each statement separately; CONCURRENTLY must not run in a transaction.
   exists=out(['docker','exec','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-Atc',"SELECT count(*) FROM pg_class WHERE relname='idx_usage_logs_api_key_created_cost'"]).strip()
   assert exists=='0','index already exists without migration record; inspect before retry'
   body="SET lock_timeout='1s'; SET statement_timeout='120s';\n"+sql+f"\nINSERT INTO schema_migrations(filename,checksum) VALUES ('{name}','{checksum}');"
  else:
   body=f"BEGIN; SET LOCAL lock_timeout='1s'; SET LOCAL statement_timeout='10s'; SELECT pg_advisory_xact_lock(694208311321144027);\n{sql}\nINSERT INTO schema_migrations(filename,checksum) VALUES ('{name}','{checksum}'); COMMIT;"
  run(['docker','exec','-i','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-v','ON_ERROR_STOP=1'],input=body.encode())
 rows=out(['docker','exec','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-Atc',"SELECT count(*) FROM pg_index WHERE indexrelid='idx_usage_logs_api_key_created_cost'::regclass AND indisvalid"]);assert rows.strip()=='1'
 a=['docker','create','--name',NEW,'--restart','unless-stopped','--security-opt','no-new-privileges:true','--ulimit','nofile=100000:100000','--env-file',str(BACKUP/'active.env'),'--network','public-proxy','--label','com.jokerdeck.release=v0.2.23-20261119']
 for m in x['Mounts']:
  assert m['Type']=='bind';a+=['-v',m['Source']+':'+m['Destination']+('' if m['RW'] else ':ro')]
 a+=[IMAGE];run(a)
 for network in x['NetworkSettings']['Networks']:
  if network!='public-proxy':run(['docker','network','connect',network,NEW])
 run(['docker','start',NEW])
 for _ in range(90):
  y=inspect(NEW);assert y['State']['Running'] and y['RestartCount']==0
  if y['State'].get('Health',{}).get('Status')=='healthy':break
  time.sleep(2)
 else:raise RuntimeError('candidate not healthy; traffic remains old')
 y=inspect(NEW);assert dict(e.split('=',1) for e in y['Config']['Env'])==dict(e.split('=',1) for e in x['Config']['Env']),'production environment drift'
 assert out(['docker','exec',NEW,'sha256sum','/app/sub2api']).split()[0]==expected
 for path,digest in resource_manifest.items():assert out(['docker','exec',NEW,'sha256sum','/app/'+path]).split()[0]==digest,('image resource drift',path)
 health(NEW);assert CONF.read_bytes()==(BACKUP/'nginx.conf').read_bytes()
 print('Candidate healthy, environment/binary verified; no traffic switched.',flush=True)
 print(out(['docker','exec',NEW,'/app/sub2api','--version']),flush=True)

def assert_public_settings_compatible(before,after):
 a,b=json.loads(before),json.loads(after);da,db=a.get('data',a),b.get('data',b)
 assert da.pop('version')=='0.2.21' and db.pop('version')=='0.2.23','unexpected release version'
 assert a==b,'existing public/custom settings changed'

def verify():
 bearer=token();a=base(OLD);b=base(NEW);report=[]
 paths=['/health','/api/v1/settings/public','/api/v1/user/profile','/api/v1/client/node-policy','/api/v1/client/node-subscription','/api/v1/admin/settings/client-node-policy','/api/v1/admin/settings/client-provider-policy','/api/v1/admin/settings/client-nodes','/api/v1/admin/settings/commerce-links','/api/v1/admin/settings/canvas-network','/api/v1/canvas/network-config','/api/v1/finance/users','/api/v1/finance/users?start_month=2026-08&end_month=2026-10','/api/v1/finance/balances','/api/v1/finance/groups','/api/v1/recharge-terms','/api/v1/self-channel-monitors','/api/v1/xianyu-recharge-orders','/api/v1/feedback/tickets','/api/v1/conversation-logs/access','/api/v1/conversation-logs','/api/v1/canvas','/api/v1/admin/announcements','/api/v1/admin/payment/config','/api/v1/admin/groups','/api/v1/admin/settings/admin-api-keys']
 strict={'/api/v1/client/node-policy','/api/v1/client/node-subscription','/api/v1/admin/settings/client-node-policy','/api/v1/admin/settings/client-provider-policy','/api/v1/admin/settings/client-nodes','/api/v1/admin/settings/commerce-links','/api/v1/admin/settings/canvas-network','/api/v1/finance/groups','/api/v1/recharge-terms','/api/v1/canvas'}
 for path in paths:
  ca,ba,_=req(a+path,bearer);cb,bb,_=req(b+path,bearer)
  assert ca==cb==200,(path,ca,cb)
  equal=ba==bb
  if path in strict:assert equal,('custom behavior changed',path)
  if path=='/api/v1/settings/public':
   assert_public_settings_compatible(ba,bb)
  report.append({'path':path,'old':ca,'new':cb,'identical':equal,'old_sha256':hashlib.sha256(ba).hexdigest(),'new_sha256':hashlib.sha256(bb).hexdigest()});print(path,ca,cb,'identical='+str(equal),flush=True)
 for path in ['/','/login','/canvas','/canvas-optimized','/admin/settings?tab=canvasRoute','/infinite-canvas/','/canvas-guide/01-canvas-home.png']:
  c,raw,h=req(b+path);assert c==200,(path,c)
  if path in ['/','/infinite-canvas/']:
   for value in re.findall(r'(?:src|href)="([^"?#]+\.(?:js|css))',raw.decode()):
    p=value if value.startswith('/') else '/infinite-canvas/'+value
    code,content,headers=req(b+p);assert code==200 and 'html' not in headers.get('Content-Type','') and content,(p,code)
 # Ensure the custom lazy UI chunks and canvas assets remain part of the new build.
 code,index,_=req(b+'/');main=re.search(r'src="([^"?#]+\.js)',index.decode()).group(1);source=req(b+main)[1].decode()
 for marker in ['OptimizedCanvasView-','FinanceView-']:assert marker in source,marker
 for path in ['/v1/messages','/v1/chat/completions','/v1/responses']:
  assert req(b+path,data=b'{}')[0]==401,path
 (STAGE/'verification.json').write_text(json.dumps(report,indent=2));print('Authenticated custom API and frontend checks passed; no purchasing/recharge/upstream calls sent.',flush=True)

def assets():
 root=pathlib.Path('/var/www/jokerdeck/client-site/upgrade-compat-v0218');a=base(OLD);b=base(NEW);queue=['/','/infinite-canvas/'];seen=set();saved=[]
 while queue:
  path=queue.pop()
  if path in seen:continue
  seen.add(path);code,body,headers=req(a+path);assert code==200,(path,code)
  kind=headers.get('Content-Type','')
  if path.endswith(('.js','.css')):
   if 'html' in kind:continue
   nc,_,nh=req(b+path)
   if nc!=200 or 'html' in nh.get('Content-Type',''):
    assert path.startswith(('/assets/','/infinite-canvas/assets/')) and '..' not in path
    target=root/path.lstrip('/');target.parent.mkdir(parents=True,exist_ok=True)
    directory=target.parent
    while directory!=root:
     directory.chmod(0o755);directory=directory.parent
    if target.exists():assert target.read_bytes()==body,('immutable asset mismatch',path)
    else:target.write_bytes(body);target.chmod(0o644)
    saved.append(path)
  if 'html' not in kind and not any(t in kind for t in ('javascript','css')):continue
  for value in re.findall(r'''["']([^"'\s]+\.(?:js|css))["']''',body.decode(errors='replace')):
   if value.startswith(('http:','https:','data:','node:')):continue
   if value.startswith('/'):p=value
   elif value.startswith('assets/'):p=('/infinite-canvas/' if path.startswith('/infinite-canvas/') else '/')+value
   elif value.startswith('.') or '/' not in value:p=urllib.parse.urljoin(path,value)
   else:continue
   if p.startswith(('/assets/','/infinite-canvas/assets/')) and p not in seen:queue.append(p)
 for path in saved:assert curl_status(HOSTS[0],path,origin=True)==200,path
 (STAGE/'old-assets.json').write_text(json.dumps(sorted(saved),indent=2));print('Assets checked=',len(seen),'old hashes preserved and verified=',len(saved),flush=True)

def switch():
 assert (STAGE/'verification.json').is_file() and (STAGE/'old-assets.json').is_file()
 raw=CONF.read_bytes();assert raw==(BACKUP/'nginx.conf').read_bytes() and raw.count(OLD_LINE)==UPSTREAMS
 assert NEW_LINE not in raw;health(OLD);health(NEW)
 for host in HOSTS:assert curl_status(host)==200 and curl_status(host,origin=True)==200
 pids=workers();assert pids
 (STAGE/'pre-switch-workers.json').write_text(json.dumps(sorted(pids)))
 timestamp=datetime.datetime.now(datetime.timezone.utc).isoformat();(STAGE/'switch-time.txt').write_text(timestamp)
 try:
  edge_reload(raw.replace(OLD_LINE,NEW_LINE))
  for _ in range(10):
   for host in HOSTS:assert curl_status(host)==200 and curl_status(host,origin=True)==200
   time.sleep(1)
  for path in ['/v1/messages','/v1/chat/completions','/v1/responses']:assert curl_status('api.jokere.asia',path,data='{}')==401
  for path in ['/','/infinite-canvas/','/api/v1/settings/public']:assert curl_status(HOSTS[0],path)==200
  assert curl_status(HOSTS[0],'/api/v1/admin/settings/client-nodes')==401
 except BaseException:
  health(OLD);edge_reload(raw)
  (STAGE/'switch-failed.txt').write_text('Rolled back Nginx; BOTH apps retained until workers drain.');raise
 print('All 11 upstreams switched by graceful reload; 60 origin/public health checks and gateway probes passed.',flush=True)

def log_report():
 logs=subprocess.run(['docker','logs','--since',(STAGE/'switch-time.txt').read_text().strip(),'edge-nginx'],capture_output=True,text=True,check=True)
 statuses=collections.Counter();gate=collections.Counter();gateway_upstream_errors=0
 for line in (logs.stdout+'\n'+logs.stderr).splitlines():
  m=re.search(r'"([A-Z]+) ([^ ]+) HTTP/[^\"]+" (\d{3}) ',line)
  if m:
   statuses[m[3]]+=1
   if m[2].startswith(('/v1/','/v1beta/','/api/v1/gateway/')) and 'Jokerdeck-Upgrade-Check/0.2.23' not in line:gate[m[3]]+=1
  if re.search(r'connect\(\) failed|no live upstreams|host not found in upstream|upstream prematurely closed connection|upstream timed out',line):gateway_upstream_errors+=1
 errors=subprocess.run(['docker','logs','--since',(STAGE/'switch-time.txt').read_text().strip(),NEW],capture_output=True,text=True,check=True)
 fatal=[x for x in (errors.stdout+'\n'+errors.stderr).splitlines() if re.search(r'(?i)panic:|fatal|out of memory|\"level\":\s*\"error\"|\tERROR\t',x)]
 return statuses,gate,gateway_upstream_errors,fatal

def observe(seconds):
 records=[];start=out(['docker','inspect','edge-nginx','--format','{{.State.StartedAt}}']).strip()
 stop=time.monotonic()+seconds;failure=None
 with (STAGE/'monitor.jsonl').open('a') as f:
  while time.monotonic()<stop:
   row={'at':datetime.datetime.now(datetime.timezone.utc).isoformat(),'origin':{},'public':{}}
   try:
    health(NEW)
    for host in HOSTS:
     row['origin'][host]=curl_status(host,origin=True);row['public'][host]=curl_status(host)
     assert row['origin'][host]==row['public'][host]==200,(host,row)
    assert out(['docker','inspect','edge-nginx','--format','{{.State.StartedAt}}']).strip()==start,'Nginx restarted'
    stats,_,upstream_errors,fatal=log_report()
    assert not upstream_errors,'Nginx upstream transport error'
    assert stats['502']==stats['504']==0,'502/504 observed since cutover'
    assert not fatal,'candidate error/fatal log entry'
   except Exception as e:failure=str(e);row['failure']=failure
   f.write(json.dumps(row)+'\n');f.flush();records.append(row)
   if failure:
    health(OLD);edge_reload((BACKUP/'nginx.conf').read_bytes());raise RuntimeError('Automatic upstream rollback: '+failure)
   time.sleep(5)
 statuses,gate,gateway_upstream_errors,fatal=log_report()
 report={'checks':len(records)*len(HOSTS)*2,'seconds':seconds,'nginx_statuses':dict(statuses),'real_gateway_statuses':dict(gate),'nginx_upstream_errors':gateway_upstream_errors,'fatal_log_lines':len(fatal),'candidate_restarts':inspect(NEW)['RestartCount'],'candidate_health':inspect(NEW)['State']['Health']['Status'],'nginx_not_restarted':True}
 (STAGE/'observation.json').write_text(json.dumps(report,indent=2));print(json.dumps(report),flush=True)
 if statuses['502'] or statuses['504'] or gateway_upstream_errors or fatal:
  health(OLD);edge_reload((BACKUP/'nginx.conf').read_bytes());raise RuntimeError('Automatic upstream rollback after log verification failure')

def finalize():
 assert json.loads((STAGE/'observation.json').read_text())['nginx_statuses'].get('502',0)==0;health(NEW)
 assert CONF.read_bytes().count(NEW_LINE)==UPSTREAMS
 x=json.loads((BACKUP/'active-inspect.json').read_text());old_image=x['Config']['Image']
 p=ROOT/'docker-compose.yml';raw=p.read_text();assert raw.count('image: '+old_image)==raw.count('container_name: '+OLD)==1
 changed=raw.replace('image: '+old_image,'image: '+IMAGE).replace('container_name: '+OLD,'container_name: '+NEW)
 p.write_text(changed)
 p=ROOT/'current.env';text=p.read_text()
 if 'IMAGE='+old_image in text:text=text.replace('IMAGE='+old_image,'IMAGE='+IMAGE)
 else:assert text==(BACKUP/'current.env').read_text(),'current.env changed since backup'
 p.write_text(text);p.chmod(0o600)
 state={'deployed_at':datetime.datetime.now(datetime.timezone.utc).isoformat(),'image':IMAGE,'container':NEW,'previous_container':OLD,'backup_dir':str(BACKUP),'rollback_script':str(STAGE/'rollback.py'),'sha256':(STAGE/'SHA256SUMS').read_text().split()[0],'version':'0.2.23','old_container_running':'True'}
 (ROOT/'.deploy-state').write_text(''.join(f'{k}={v}\n' for k,v in state.items()));(ROOT/'.deploy-state').chmod(0o600)
 run(['docker','compose','--env-file',str(ROOT/'current.env'),'-f',str(ROOT/'docker-compose.yml'),'config','--quiet'])
 print('Compose and deployment state synchronized without recreating any production containers.',flush=True)

def retire():
 before=set(json.loads((STAGE/'pre-switch-workers.json').read_text()));remaining=before & workers()
 assert not remaining,('Old Nginx workers still serving long requests; old app MUST stay running',sorted(remaining))
 health(NEW);assert CONF.read_bytes().count(NEW_LINE)==UPSTREAMS
 run(['docker','update','--restart=no',OLD]);run(['docker','stop','-t','120',OLD])
 assert not inspect(OLD)['State']['Running'];health(NEW)
 p=ROOT/'.deploy-state';p.write_text(p.read_text().replace('old_container_running=True','old_container_running=False'));p.chmod(0o600)
 (STAGE/'retired.json').write_text(json.dumps({'old_stopped':True,'nginx_old_workers_drained':True,'old_retained':True,'restart_policy':'no','at':datetime.datetime.now(datetime.timezone.utc).isoformat()}));print('Old workers drained; previous app gracefully stopped and retained for rollback.',flush=True)

def rollback():
 run(['docker','start',OLD]);run(['docker','update','--restart=unless-stopped',OLD])
 for _ in range(90):
  if inspect(OLD)['State'].get('Health',{}).get('Status')=='healthy':break
  time.sleep(2)
 else:raise RuntimeError('Old app not healthy; Nginx left unchanged')
 health(OLD);pids=workers();raw=CONF.read_bytes()
 assert raw.count(OLD_LINE)==UPSTREAMS or raw.count(NEW_LINE)==UPSTREAMS
 edge_reload((BACKUP/'nginx.conf').read_bytes())
 for name in ['docker-compose.yml','.deploy-state','current.env']:
  (ROOT/name).write_bytes((BACKUP/name).read_bytes());(ROOT/name).chmod(0o600)
 for host in HOSTS:assert curl_status(host,origin=True)==200
 for _ in range(120):
  if not (pids & workers()):
   run(['docker','update','--restart=no',NEW]);run(['docker','stop','-t','120',NEW]);print('Rollback complete; new workers drained, candidate stopped. Additive DB schema kept.');return
  time.sleep(5)
 print('Rollback complete; candidate retained RUNNING while its long connections drain. Do not force-stop.')

if __name__=='__main__':
 phase=sys.argv[1]
 if phase=='observe':observe(int(sys.argv[2]) if len(sys.argv)>2 else 600)
 else:{'start':start,'verify':verify,'assets':assets,'switch':switch,'finalize':finalize,'retire':retire,'rollback':rollback}[phase]()
