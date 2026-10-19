// Purpose: thin local IPC bridge and tray lifecycle. Creator: HIHU-COCO. Created: 2026-10-19.
use serde_json::Value;
use std::time::Duration;
use tauri::{
    menu::{Menu, MenuItem},
    tray::TrayIconBuilder,
    Manager,
};
use tokio::io::{AsyncReadExt, AsyncWriteExt};

fn endpoint() -> Result<String, String> {
    if let Ok(value) = std::env::var("HOPNODE_POC_ENDPOINT") {
        return Ok(value);
    }
    #[cfg(windows)]
    {
        Ok(format!(
            r"\\.\pipe\hopnode-poc-{}",
            std::env::var("USERNAME").map_err(|e| e.to_string())?
        ))
    }
    #[cfg(unix)]
    {
        std::env::var("XDG_RUNTIME_DIR")
            .map(|dir| format!("{dir}/hopnode-poc.sock"))
            .map_err(|_| {
                "Set XDG_RUNTIME_DIR or HOPNODE_POC_ENDPOINT to the private controller socket"
                    .to_string()
            })
    }
}

fn parse_response(bytes: &[u8]) -> Result<Value, String> {
    let split = bytes
        .windows(4)
        .position(|v| v == b"\r\n\r\n")
        .ok_or("Invalid Core HTTP response")?;
    let header = std::str::from_utf8(&bytes[..split]).map_err(|e| e.to_string())?;
    if !header.starts_with("HTTP/1.1 200 ") && !header.starts_with("HTTP/1.0 200 ") {
        return Err(format!(
            "Core returned {}",
            header.lines().next().unwrap_or("an invalid status")
        ));
    }
    let value: Value = serde_json::from_slice(&bytes[split + 4..])
        .map_err(|e| format!("Invalid Core JSON: {e}"))?;
    if value["api_version"] != "poc.v0" {
        return Err("Incompatible Core IPC version".to_string());
    }
    Ok(value)
}

async fn query_status() -> Result<Value, String> {
    let path = endpoint()?;
    #[cfg(windows)]
    let stream = tokio::net::windows::named_pipe::ClientOptions::new()
        .open(path)
        .map_err(|e| format!("Core Named Pipe: {e}"))?;
    #[cfg(unix)]
    let stream = tokio::net::UnixStream::connect(path)
        .await
        .map_err(|e| format!("Core Unix Socket: {e}"))?;
    let mut stream = stream.take(65_537);
    stream
        .get_mut()
        .write_all(b"GET /v0/status HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
        .await
        .map_err(|e| e.to_string())?;
    let mut reply = Vec::new();
    stream
        .read_to_end(&mut reply)
        .await
        .map_err(|e| e.to_string())?;
    if reply.len() > 65_536 {
        return Err("Core response exceeded 64 KiB".to_string());
    }
    parse_response(&reply)
}

#[tauri::command]
async fn core_status() -> Result<Value, String> {
    tokio::time::timeout(Duration::from_secs(3), query_status())
        .await
        .map_err(|_| "Core IPC timed out after 3 seconds".to_string())?
}

pub fn run() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![core_status])
        .setup(|app| {
            let show = MenuItem::with_id(app, "show", "打开跃点 HopNode", true, None::<&str>)?;
            let quit = MenuItem::with_id(app, "quit", "退出桌面程序", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show, &quit])?;
            TrayIconBuilder::new()
                .icon(
                    app.default_window_icon()
                        .ok_or("Missing tray icon")?
                        .clone(),
                )
                .tooltip("跃点 HopNode")
                .menu(&menu)
                .on_menu_event(|app, event| match event.id().as_ref() {
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    "quit" => app.exit(0),
                    _ => {}
                })
                .build(app)?;
            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                api.prevent_close();
                let _ = window.hide();
            }
        })
        .run(tauri::generate_context!())
        .expect("Unable to start HopNode desktop");
}

#[cfg(test)]
mod tests {
    use super::parse_response;
    #[test]
    fn response_requires_success_and_matching_version() {
        assert!(parse_response(b"HTTP/1.1 200 OK\r\n\r\n{\"api_version\":\"poc.v0\"}").is_ok());
        assert!(parse_response(b"HTTP/1.1 403 Forbidden\r\n\r\n{}").is_err());
        assert!(parse_response(b"HTTP/1.1 200 OK\r\n\r\n{\"api_version\":\"other\"}").is_err());
        assert!(parse_response(b"invalid").is_err());
    }

    #[test]
    #[ignore = "requires a separately running P04 Go controller and HOPNODE_POC_ENDPOINT"]
    fn real_controller_roundtrip() {
        let runtime = tokio::runtime::Builder::new_current_thread()
            .enable_all()
            .build()
            .unwrap();
        let status = runtime
            .block_on(super::core_status())
            .expect("real Go controller IPC failed");
        assert_ne!(
            status["pid"].as_u64().unwrap(),
            u64::from(std::process::id())
        );
        assert!(!status["identity"].as_str().unwrap().is_empty());
        println!(
            "Real Go Core IPC PASS: api={}, version={}, distinct process=true",
            status["api_version"], status["version"]
        );
    }
}
