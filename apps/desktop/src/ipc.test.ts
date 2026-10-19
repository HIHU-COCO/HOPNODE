// Purpose: reject malformed or incompatible controller replies. Creator: HIHU-COCO. Created: 2026-10-19.
import { describe, expect, it } from 'vitest';
import { parseStatus } from './ipc';

const status = {
  api_version: 'poc.v0',
  version: '0.1.0-poc',
  pid: 123,
  instance_id: 'test',
  os: 'linux',
  identity: 'user',
  started_at: '2026-10-19T00:00:00Z',
};
describe('Core boundary', () => {
  it('preserves controller process identity', () =>
    expect(parseStatus(status)).toEqual(status));
  it('rejects incompatible IPC versions', () =>
    expect(() => parseStatus({ ...status, api_version: 'v999' })).toThrow(
      '不兼容',
    ));
  it('rejects malformed identity', () =>
    expect(() => parseStatus({ ...status, pid: 0 })).toThrow('进程标识'));
  it('rejects missing status details', () =>
    expect(() => parseStatus({ ...status, started_at: '' })).toThrow(
      'started_at',
    ));
});
