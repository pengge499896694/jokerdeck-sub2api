import pathlib,subprocess,json,os,datetime
os.umask(0o077)
app=pathlib.Path('/opt/sub2api');stage=app/'staging/v0.2.25-20261125';backup=app/'backups/pre-v0.2.25-20261125'
old='sub2api-v0.2.24-20261124';line=f'set $sub2api_upstream http://{old}:8080;'.encode()
def out(a):return subprocess.check_output(a,text=True)
x=json.loads(out(['docker','inspect',old]))[0]
assert x['State']['Running'] and x['State']['Health']['Status']=='healthy'
conf=pathlib.Path('/opt/edge/nginx.conf').read_bytes();assert conf.count(line)==11
assert subprocess.check_output(['docker','exec','edge-nginx','cat','/etc/nginx/conf.d/default.conf'])==conf
subprocess.run(['docker','exec','edge-nginx','nginx','-t'],check=True)
assert not backup.exists(), 'backup directory already exists; inspect before retry'
backup.mkdir(mode=0o700)
(backup/'active-inspect.json').write_text(json.dumps(x,indent=2));(backup/'active.env').write_text('\n'.join(x['Config']['Env'])+'\n')
(backup/'nginx.conf').write_bytes(conf)
for name in ['docker-compose.yml','.deploy-state','current.env','.env','data/config.yaml']:
 p=app/name
 if p.exists():(backup/name.replace('/','-')).write_bytes(p.read_bytes())
remote=dict(line.split('|') for line in out(['docker','exec','sub2api-postgres','psql','-U','sub2api','-d','sub2api','-At','-F','|','-c','SELECT filename,checksum FROM schema_migrations ORDER BY filename']).splitlines())
expected=json.loads((stage/'migrations.json').read_text())
assert all(expected.get(k)==v for k,v in remote.items()),'applied migration checksum drift'
added=sorted(set(expected)-set(remote));assert added==[],added
(backup/'migrations.json').write_text(json.dumps(remote,indent=2))
for name,args in [('database.dump',['pg_dump','-U','sub2api','-d','sub2api','-Fc']),('schema.sql',['pg_dump','-U','sub2api','-d','sub2api','--schema-only','--no-owner','--no-privileges'])]:
 with (backup/name).open('wb') as f:subprocess.run(['docker','exec','sub2api-postgres',*args],stdout=f,check=True)
with (backup/'database.dump').open('rb') as f:subprocess.run(['docker','exec','-i','sub2api-postgres','pg_restore','-l'],stdin=f,stdout=subprocess.DEVNULL,check=True)
(stage/'backup-time.txt').write_text(datetime.datetime.now(datetime.timezone.utc).isoformat())
print(json.dumps({'backup':str(backup),'dump_bytes':(backup/'database.dump').stat().st_size,'applied_migrations_verified':len(remote),'added':added,'old_serving':old}))
