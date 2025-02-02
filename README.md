## О этом проекте

Это сборка [Minekube Gate](https://github.com/minekube/gate-plugin-template) с плагинами для серверов Auranite.

### Плагин supaauth
Добавляет возможность авторизации через Supabase

```
# Создать конфиг файл в корне supaauth.yml
enable: false # Включает плагин
enableIpWhitelist: false # Включает фильтр по IP
supabaseApiKey: apiKey # Апи ключ с Supabase
supabaseApiUrl: apiUrl # Апи урл с Supabase
changePlayerUUID: false # Меняет uuid игрока на uuid с Supabase
changePlayerUsername: false # Меняет имя игрока на имя с Supabase
```