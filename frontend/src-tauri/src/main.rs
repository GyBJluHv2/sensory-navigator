// Запрещаем создание дополнительного консольного окна на Windows в release-сборке
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    sensory_navigator_lib::run()
}
