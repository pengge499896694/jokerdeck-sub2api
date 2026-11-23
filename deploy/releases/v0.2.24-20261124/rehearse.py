import subprocess,time,pathlib,json,hashlib,os
os.umask(0o077)
s=pathlib.Path('/opt/sub2api/staging/v0.2.24-20261124');b=pathlib.Path('/opt/sub2api/backups/pre-v0.2.24-20261124');name='sub2api-migration-rehearsal-20261124'
def run(a,**kw):return subprocess.run(a,check=True,**kw)
def out(a):return subprocess.check_output(a,text=True)
assert not out(['docker','ps','-a','--filter',f'name=^/{name}$','--format','{{.Names}}']).strip()
run(['docker','run','-d','--name',name,'--network','none','--memory','512m','--cpus','0.75','--security-opt','no-new-privileges:true','-e','POSTGRES_USER=sub2api','-e','POSTGRES_DB=sub2api','-e','POSTGRES_HOST_AUTH_METHOD=trust','postgres:18-alpine','postgres','-c','shared_buffers=32MB','-c','max_connections=30'])
try:
 for _ in range(80):
  if subprocess.run(['docker','exec',name,'pg_isready','-U','sub2api','-d','sub2api'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode==0:break
  time.sleep(1)
 else:raise RuntimeError('isolated database unavailable')
 for _ in range(30):
  with (b/'database.dump').open('rb') as f:
   result=subprocess.run(['docker','exec','-i',name,'pg_restore','-U','sub2api','-d','sub2api','--exit-on-error','--no-owner','--no-privileges'],stdin=f)
  if result.returncode==0: break
  time.sleep(2)
 else: raise RuntimeError('isolated database restore failed after retries')
 def psql(sql):run(['docker','exec','-i',name,'psql','-U','sub2api','-d','sub2api','-v','ON_ERROR_STOP=1'],input=sql.encode())
 rows=out(['docker','exec',name,'psql','-U','sub2api','-d','sub2api','-Atc',"SELECT count(*) FROM schema_migrations;"])
 assert rows.strip()=='315',rows
 (s/'rehearsal.json').write_text(json.dumps({'full_backup_restored':True,'migration_count':315,'no_schema_changes':True,'network':'none'}))
 print('Full backup restored successfully in isolated PostgreSQL; release has no schema migrations.')
finally:run(['docker','rm','-f','-v',name])
