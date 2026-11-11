"""Offline safety checks for the topology-specific release; no Docker/network calls."""
import collections
import contextlib
import io
import json
import pathlib
import subprocess
import tempfile
import unittest
from unittest.mock import Mock, patch

SOURCE = pathlib.Path(__file__).with_name('deploy.py')


def load():
    ns = {'__name__': 'release_test', '__file__': str(SOURCE)}
    exec(compile(SOURCE.read_text(), str(SOURCE), 'exec'), ns)
    return ns


class ReleaseSafetyTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.ns = load()
        self.root = pathlib.Path(self.tmp.name)
        for key in ('ROOT', 'STAGE', 'BACKUP'):
            self.ns[key] = self.root / key.lower()
            self.ns[key].mkdir()
        self.ns['CONF'] = self.root / 'nginx.conf'
        self.ns['health'] = Mock()
        self.ns['run'] = Mock()
        self.ns['out'] = Mock(return_value='unchanged-start-time')

    def test_reload_keeps_bind_mount_inode_and_only_reloads(self):
        conf = self.ns['CONF']
        conf.write_bytes(b'old')
        inode = conf.stat().st_ino
        self.ns['out'].return_value = 'new'
        with patch.object(subprocess, 'check_output', return_value=b'new'):
            self.ns['edge_reload'](b'new')
        self.assertEqual(conf.stat().st_ino, inode)
        self.assertEqual(conf.read_bytes(), b'new')
        commands = [c.args[0] for c in self.ns['run'].call_args_list]
        self.assertEqual(commands, [
            ['docker', 'exec', 'edge-nginx', 'nginx', '-t'],
            ['docker', 'exec', 'edge-nginx', 'nginx', '-s', 'reload'],
        ])

    def test_finalize_accepts_application_env_without_image_key(self):
        old = self.ns['OLD']
        (self.ns['BACKUP'] / 'active-inspect.json').write_text('{"Config":{"Image":"old-image"}}')
        (self.ns['STAGE'] / 'observation.json').write_text('{"nginx_statuses":{"200":20}}')
        (self.ns['STAGE'] / 'SHA256SUMS').write_text('abcdef  sub2api\n')
        (self.ns['ROOT'] / 'docker-compose.yml').write_text(f'services:\n  app:\n    image: old-image\n    container_name: {old}\n')
        env = 'JWT_SECRET=not-a-real-secret\nSERVER_PORT=8080\n'
        for folder in ('ROOT', 'BACKUP'):
            (self.ns[folder] / 'current.env').write_text(env)
        self.ns['CONF'].write_bytes(self.ns['NEW_LINE'] * self.ns['UPSTREAMS'])
        with contextlib.redirect_stdout(io.StringIO()):
            self.ns['finalize']()
        self.assertEqual((self.ns['ROOT'] / 'current.env').read_text(), env)
        self.assertIn(self.ns['NEW'], (self.ns['ROOT'] / 'docker-compose.yml').read_text())
        self.assertIn('old_container_running=True', (self.ns['ROOT'] / '.deploy-state').read_text())
        self.assertEqual(len(self.ns['run'].call_args_list), 1)  # compose config only, no recreation

    def test_log_report_separates_probes_and_detects_transport_errors(self):
        (self.ns['STAGE'] / 'switch-time.txt').write_text('2026-11-11T00:00:00Z')
        logs = '\n'.join([
            'x "POST /v1/responses HTTP/1.1" 200 120 "-" "client-sdk"',
            'x "POST /v1/responses HTTP/1.1" 401 120 "-" "Jokerdeck-Upgrade-Check/0.2.20"',
            'x "GET /health HTTP/2.0" 502 120 "-" "curl"',
            'connect() failed (111: Connection refused) while connecting to upstream',
        ])
        responses = [subprocess.CompletedProcess([], 0, logs, ''), subprocess.CompletedProcess([], 0, '', '')]
        with patch.object(subprocess, 'run', side_effect=responses):
            statuses, real, transport, fatal = self.ns['log_report']()
        self.assertEqual(statuses['502'], 1)
        self.assertEqual(real, {'200': 1})
        self.assertEqual(transport, 1)
        self.assertEqual(fatal, [])

    def test_end_of_observation_failure_rolls_back(self):
        (self.ns['BACKUP'] / 'nginx.conf').write_bytes(b'old-upstream')
        self.ns['inspect'] = Mock(return_value={'RestartCount': 0, 'State': {'Health': {'Status': 'healthy'}}})
        self.ns['log_report'] = Mock(return_value=(collections.Counter({'502': 1}), collections.Counter(), 0, []))
        self.ns['edge_reload'] = Mock()
        with contextlib.redirect_stdout(io.StringIO()), self.assertRaisesRegex(RuntimeError, 'Automatic upstream rollback'):
            self.ns['observe'](0)
        self.ns['health'].assert_called_with(self.ns['OLD'])
        self.ns['edge_reload'].assert_called_once_with(b'old-upstream')
        self.assertTrue((self.ns['STAGE'] / 'observation.json').exists())

    def test_public_settings_only_allow_version_and_disabled_mcp_addition(self):
        old = {'data': {'version': '0.2.19', 'payment_enabled': True, 'canvas': {'relay': 'unchanged'}}}
        new = {'data': dict(old['data'], version='0.2.20', mcp_enabled=False)}
        check = self.ns['assert_public_settings_compatible']
        check(json.dumps(old), json.dumps(new))
        new['data']['mcp_enabled'] = True
        with self.assertRaises(AssertionError):
            check(json.dumps(old), json.dumps(new))
        new['data']['mcp_enabled'] = False
        new['data']['canvas'] = {'relay': 'changed'}
        with self.assertRaises(AssertionError):
            check(json.dumps(old), json.dumps(new))
        new['data']['canvas'] = old['data']['canvas']
        new['data']['extra_setting'] = False
        with self.assertRaises(AssertionError):
            check(json.dumps(old), json.dumps(new))

    def test_old_workers_prevent_retiring_previous_app(self):
        (self.ns['STAGE'] / 'pre-switch-workers.json').write_text('[12,13]')
        self.ns['workers'] = Mock(return_value={12, 42})
        with self.assertRaisesRegex(AssertionError, 'MUST stay running'):
            self.ns['retire']()
        self.ns['run'].assert_not_called()

    def test_successful_retirement_is_graceful_and_retains_container(self):
        (self.ns['STAGE'] / 'pre-switch-workers.json').write_text('[12,13]')
        (self.ns['ROOT'] / '.deploy-state').write_text('old_container_running=True\n')
        self.ns['CONF'].write_bytes(self.ns['NEW_LINE'] * self.ns['UPSTREAMS'])
        self.ns['workers'] = Mock(return_value={42, 43})
        self.ns['inspect'] = Mock(return_value={'State': {'Running': False}})
        with contextlib.redirect_stdout(io.StringIO()):
            self.ns['retire']()
        commands = [c.args[0] for c in self.ns['run'].call_args_list]
        self.assertIn(['docker', 'stop', '-t', '120', self.ns['OLD']], commands)
        self.assertFalse(any('rm' in cmd or 'kill' in cmd for cmd in commands))


if __name__ == '__main__':
    unittest.main()
