### GoNews

Сервис агрегатор новостных статей из RSS лент. Практика на курсе "Go-разработчик" от SkillFactory. Часть итогового проекта курса.

Для запуска нужно установить путь к файлу конфига в переменную окружения `NEWS_CONFIG_PATH`, пароль для доступа к MongoDB
в переменную окружения `NEWS_DB_PASSWD`. Остальные входные данные указываются в файле конфига.

Сам файл конфига `config.yaml` лежит в каталоге config.

**Сделано:**

- Использование базы данных MongoDB с настроенной авторизацией.
- Логирование в stdout через пакет slog стандартной библиотеки Go.
- Парсинг RSS лент с указанных в `config.yaml` адресов и запись полученных новостных статей в базу данных.
- REST API метод возврата указанного количества последних по дате публикации новостных статей из базы данных.
- Эмуляция базы данных в памяти для облегчения тестирования. НЕ ИСПОЛЬЗУЕТСЯ.
- Эмуляция внешних ресурсов (RSS ленты сайта, базы данных) через генерацию моков из библиотеки Mockery.
- Тесты для всех основных пакетов приложения.
- Использование контекстов при работе парсера, сервера и базы данных.
- Завершение работы приложения по сигналу прерывания с использованием graceful shutdown.
