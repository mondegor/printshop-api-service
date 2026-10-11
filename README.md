# Описание Print Shop Back v0.18.0
Этот репозиторий содержит описание сервиса Print Shop Back.

## Статус сервиса
Сервис находится в стадии разработки.

## Описание сервиса
Web сервис для расчёта стоимости и времени изготовления продукции.

Подробнее смотри [документацию проекта](./docs/README.md)

> Перед запуском консольных скриптов сервиса необходимо скачать и установить утилиту Mrcmd.\
> Инструкция по её установке находится [здесь](https://github.com/mondegor/mrcmd#readme)

## Разворачивание, установка и запуск сервиса

### Разворачивание сервиса
- Выбрать рабочую директорию, где должен быть расположен сервис
- `mkdir print-shop-back && cd print-shop-back` // создать и перейти в директорию проекта
- `git clone git@github.com:mondegor/print-shop-back.git .`
- `cp .env.dist-stand .env`
- `mrcmd state` // проверка состояния сервиса
- `mrcmd config` // проверка установленных переменных сервиса
- `mrcmd go deps` // загрузка зависимостей проекта
- Для работы утилит `golangci-lint` (линтеры и форматтеры), `mockgen` необходимо в `.env` проверить
  значения переменных `GO_TOOLS_INSTALL_*` и запустить `mrcmd go install-tools`. Версии
  `golangci-lint` (та же, что в CI) и `mockgen` заданы в `.env.dist`;

### Установка сервиса и его первый запуск
- `mrcmd docker ps` // убеждаемся, что Docker daemon запущен
- `mrcmd install`
- `mrcmd start`
- `mrcmd docker-compose ps` // проверка всех запущенных ресурсов сервиса
- `mrcmd go logs` // проверка запуска основного сервиса

### Запуск и остановка сервиса
- `mrcmd start`
- `mrcmd stop`

### Остановка сервиса и удаление всех его установленных ресурсов
- `mrcmd uninstall`

### Запуск тестов
Тесты сервиса запускаются из директории `app` командой `go test ./...`. Интеграционные тесты
(`app/tests/integration`, тесты репозиториев `*_postgres_test.go`) поднимают Postgres и Redis через
`testcontainers` и требуют запущенный Docker, без него `go test ./...` завершится с ошибкой.

### Часто используемые команды
- `mrcmd help` // помощь в контексте текущего сервиса;
- `mrcmd state` // общее состояние текущего сервиса;
- `mrcmd docker-compose conf` // отображает список `.yaml` файлов из которых собрана конфигурация и саму конфигурацию;
- `mrcmd docker-compose ps` // текущее состояние запущенных ресурсов;
- `mrcmd docker-compose logs` // логи всех запущенных ресурсов;
- `mrcmd go-migrate help` // все команды сервиса go-migrate;
- `mrcmd go-migrate up` // загрузка дампа с данными в БД;
- `mrcmd postgres help` // все команды сервиса postgres;
- `mrcmd go help` // выводит список всех доступных go команд;
- `mrcmd go generate` // генерирует go файлы через встроенный механизм go:generate;
- `mrcmd go fmt` // исправляет форматирование кода и imports (`golangci-lint fmt`: форматтеры из раздела `formatters` в `.golangci.yaml`, сгенерированные файлы пропускаются);
- `mrcmd go fmt-diff` // показывает изменения, которые внесёт `fmt`, не меняя файлы (`golangci-lint fmt -d`);
- `mrcmd go lint` // запускает линтеры для проверки кода (на основе `.golangci.yaml`);
- `mrcmd go lint-verify` // проверяет `.golangci.yaml` по JSON-схеме (`golangci-lint config verify`, нужен доступ к golangci-lint.run);
- `mrcmd go test` // запускает тесты сервиса;
- `mrcmd go test-report` // запускает тесты сервиса с формированием отчёта о покрытии кода (`test-coverage-full.html`);
- `mrcmd plantuml build-all` // генерирует файлы изображений из `.puml` [подробнее](https://github.com/mondegor/mrcmd-plugins/blob/master/plantuml/README.md#%D1%80%D0%B0%D0%B1%D0%BE%D1%82%D0%B0-%D1%81-%D0%B4%D0%BE%D0%BA%D1%83%D0%BC%D0%B5%D0%BD%D1%82%D0%B0%D1%86%D0%B8%D0%B5%D0%B9-%D0%BF%D1%80%D0%BE%D0%B5%D0%BA%D1%82%D0%B0-markdown--plantuml);

> Более подробную информацию по использованию утилиты Mrcmd
> смотрите [здесь](https://github.com/mondegor/mrcmd#readme).

#### Короткий вариант выше приведённых команд (Makefile)
- `make build` // аналог `mrcmd install`
- `make deps` // аналог `mrcmd go deps`
- `make deps-upgrade` // аналог `mrcmd go get -u ./...` + `mrcmd go tidy`
- `make migrate` // аналог `mrcmd go-migrate up`
- `make generate` // аналог `mrcmd go generate`
- `make lint` // аналог `mrcmd go fmt` + `lint`
- `make lint-verify` // аналог `mrcmd go lint-verify`
- `make test` // аналог `mrcmd go test`
- `make test-report` // аналог `mrcmd go test-report`
- `make plantuml` // аналог `mrcmd plantuml build-all`
- `make app-conf` // аналог `mrcmd config && mrcmd docker-compose conf`
- `make app-start` // аналог `mrcmd start`
- `make app-state` // аналог `mrcmd docker-compose ps`
- `make app-logs` // аналог `mrcmd docker-compose logs`
- `make app-stop` // аналог `mrcmd stop`

> Чтобы расширить список команд, необходимо создать Makefile.mk и добавить
> туда дополнительные команды, все они будут добавлены в единый список команд make утилиты.

## Панели управления развёрнутой инфраструктуры
- TRAEFIK: http://traefik.local/ (admin 12345678);
- MINIO: http://minio.local/ (admin 12345678);
- API: http://api.print-shop.local/;
- HEALTH: http://print-shop.internal/health;
- INFO: http://print-shop.internal/v1/system-info;
- METRICS: http://print-shop.internal/metrics;

### Использование локальных доменов
Необходимо в hosts добавить следующие записи:
- `127.0.0.1 traefik.local`
- `127.0.0.1 minio.local`
- `127.0.0.1 print-shop.local`
- `127.0.0.1 api.print-shop.local`
- `127.0.0.1 print-shop.internal`
