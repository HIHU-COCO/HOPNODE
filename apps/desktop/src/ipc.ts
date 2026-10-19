// Purpose: typed, validated boundary to the thin Tauri IPC bridge. Creator: HIHU-COCO. Created: 2026-10-19.
import { invoke, isTauri } from '@tauri-apps/api/core';

export interface CoreStatus {
  api_version: 'poc.v0';
  version: string;
  pid: number;
  instance_id: string;
  os: string;
  identity: string;
  started_at: string;
}

export function parseStatus(input: unknown): CoreStatus {
  if (!input || typeof input !== 'object')
    throw new Error('Core 返回了无效状态。');
  const value = input as Record<string, unknown>;
  if (value.api_version !== 'poc.v0') throw new Error('Core IPC 版本不兼容。');
  if (!Number.isInteger(value.pid) || (value.pid as number) <= 0)
    throw new Error('Core 进程标识无效。');
  for (const key of [
    'version',
    'instance_id',
    'os',
    'identity',
    'started_at',
  ]) {
    if (typeof value[key] !== 'string' || value[key] === '')
      throw new Error(`Core 缺少字段：${key}`);
  }
  return value as unknown as CoreStatus;
}

export async function getCoreStatus(): Promise<CoreStatus> {
  if (!isTauri())
    throw new Error(
      '请通过 Tauri 桌面程序读取本机 Core；浏览器预览无法访问本机 IPC。',
    );
  return parseStatus(await invoke<unknown>('core_status'));
}
