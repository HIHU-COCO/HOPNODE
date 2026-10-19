// Purpose: P04 real controller diagnostics, before business UI work. Creator: HIHU-COCO. Created: 2026-10-19.
import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  ConfigProvider,
  Descriptions,
  Layout,
  Space,
  Tag,
  Typography,
  theme,
} from 'antd';
import { getCoreStatus, type CoreStatus } from './ipc';

export function App() {
  const [status, setStatus] = useState<CoreStatus>();
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(false);
  const refresh = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      setStatus(await getCoreStatus());
    } catch (cause) {
      setStatus(undefined);
      setError(String(cause));
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#55d6bb',
          borderRadius: 12,
          fontFamily: 'Segoe UI, system-ui, sans-serif',
        },
      }}
    >
      <Layout className="shell">
        <Layout.Header className="header">
          <Space size={16}>
            <span className="brand-mark">H</span>
            <Typography.Title level={3}>跃点 HopNode</Typography.Title>
            <Tag color="cyan">工程预览 · P04</Tag>
          </Space>
        </Layout.Header>
        <Layout.Content className="content">
          <Typography.Title level={2}>本机连接诊断</Typography.Title>
          <Typography.Paragraph type="secondary">
            跨越边界，连接无限。此工程预览验证桌面窗口到独立 Go Core
            的真实本机通信。
          </Typography.Paragraph>
          <Card
            title={
              <Space>
                Controller{' '}
                <Tag
                  color={status ? 'success' : error ? 'error' : 'processing'}
                >
                  {status ? '已连接' : error ? '未连接' : '正在检查'}
                </Tag>
              </Space>
            }
            extra={
              <Button
                onClick={() => {
                  void refresh();
                }}
                loading={loading}
              >
                刷新状态
              </Button>
            }
          >
            {error && (
              <Alert
                type="error"
                showIcon
                title="无法连接本机 Core"
                description={error}
              />
            )}
            {status && (
              <Descriptions
                column={{ xs: 1, md: 2 }}
                items={[
                  {
                    key: 'version',
                    label: 'Core 版本',
                    children: status.version,
                  },
                  {
                    key: 'api',
                    label: 'IPC 版本',
                    children: status.api_version,
                  },
                  { key: 'os', label: '操作系统', children: status.os },
                  {
                    key: 'identity',
                    label: '运行身份',
                    children: status.identity,
                  },
                  { key: 'pid', label: '独立进程 PID', children: status.pid },
                  {
                    key: 'started',
                    label: '启动时间',
                    children: status.started_at,
                  },
                  {
                    key: 'instance',
                    label: '实例标识',
                    children: status.instance_id,
                    span: 2,
                  },
                ]}
              />
            )}
          </Card>
          <Alert
            className="notice"
            type="info"
            showIcon
            title="当前预览范围"
            description="本页只展示真实 Core 版本与进程状态。设备、SSH、托管受控、文件、终端及 Codex 接入尚未进入业务开发。关闭窗口会隐藏到托盘；退出桌面程序不会停止独立启动的 Core。"
          />
        </Layout.Content>
      </Layout>
    </ConfigProvider>
  );
}
