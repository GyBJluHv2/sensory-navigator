// Конфигурация Tauri для приложения "Сенсорный навигатор".
// Подключает плагин shell для безопасного открытия внешних ссылок (OpenStreetMap)
// и регистрирует команду app_info, доступную из фронтенда.

#[derive(serde::Serialize)]
struct AppInfo {
    name: &'static str,
    version: &'static str,
}

#[tauri::command]
fn app_info() -> AppInfo {
    AppInfo {
        name: "Сенсорный навигатор",
        version: env!("CARGO_PKG_VERSION"),
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![app_info])
        .run(tauri::generate_context!())
        .expect("ошибка запуска Tauri-приложения");
}
