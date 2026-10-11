# Print Shop Back Changelog
Все изменения сервиса Print Shop Back будут документироваться на этой странице.

## 2026-10-11
### Changed
- Форматирование кода переведено на `golangci-lint fmt` (форматтеры и их настройки берутся
  из раздела `formatters` в `.golangci.yaml`, сгенерированные файлы пропускаются):
  `make lint` вызывает `mrcmd go fmt` вместо `fmt`, `fmti`, `fmti2`;
  для просмотра изменений без записи добавлена команда `mrcmd go fmt-diff`;
- Из `.env.dist` и `.env.dist-stand` удалены переменные `GO_IMPORTS_LOCAL_PREFIXES` и
  `GO_TOOLS_INSTALL_{GOFUMPT,GOIMPORTS,GCI}_VERSION`, из `app/.env.dist` — их аналоги `GO_DEV_*`
  (`app/Makefile` тоже переведён на `mrcmd go-dev fmt`): утилиты `gofumpt`, `goimports`, `gci`
  больше не нужны, локальный префикс imports задаётся только в `.golangci.yaml`;
- В `.env.dist` закреплена версия `golangci-lint` v2.11.4 — та же, что в CI;
  в уже созданном `.env` обновите `GO_TOOLS_INSTALL_GOLANGCI_LINT_VERSION` вручную;

### Fixed
- При переносе ровно половины остатка контейнера (`MoveStockContainer`, например 5 из 10) на новое
  место записывалось всё исходное количество, а у старого места оставалась половина: признак частичного
  переноса сравнивался с уже уменьшенным остатком. Теперь на новое место всегда уходит переносимое количество

### Added
- Тест юзкейса `AddMoreStockContainer` (warehousing) на моках `gomock`: успешное добавление
  с расчётом объёма контейнера, проверка входных данных и обработка ошибок хранилищ,
  проверки склада и его пересчёта; моки генерируются в `usr/usecase/mock`;
- Тест юзкейса `MoveStockContainer` на моках: частичный и полный перенос (включая перенос ровно половины),
  перенос на текущее место, ограничения вложения и обработка ошибок хранилища и пересчёта локаций

## 2025-03-07
### Added
- Добавление алгоритма вариантов спуска полос;
- Добавление алгоритма разложения листов в пачки;

### Changed
- Доработка алгоритма спуск полос;
- Доработка алгоритма разложения изделий в коробки;
- Рефакторинг библиотеки алгоритмов;

## 2024-12-10
### Fixed
- Исправлена неправильная работа pub методов:
    - /v1/catalog/papers/*;
    - /v1/catalog/laminates/*;

## 2024-12-01
### Changed
- Настройки для отправки писем и телеграмм сообщений теперь необязательны;

### Fixed
- В списке печатных форматов поля размеров приведены к метрам;

## 2024-11-30
### Added
- Добавлены настройки для отправки писем и телеграмм сообщений;

### Changed
- Обновлен `debugging.error_caller.upper_bounds`;
- Заменён `UseCaseHelper` -> `UseCaseErrorWrapper`;

## 2024-11-17
### Added
- Подключены сервисы `mrnotifier` и `mrmailer` для отправки персонализированных уведомлений;
- В конфиг добавлен раздел `validation`, в который вынесены все настройки по загрузке файлов и изображений, ограничения по типам, размерам;
- Внедрены служебные процессы `signal.Interception` и `onstartup.Process`;
- Повсеместно внедрён `decorator.SourceEmitter`;
- Подключён `logger.NewQueryTracer` для `Postgres`;
- Подключён и настроен линтер `gci`;
- В метод /v1/dictionaries/print-formats каждому элементу добавлены поля `Width` и `Height`;

### Changed
- Внедрена новая библиотека формирования SQL запросов, заменён дублирующий код в репозиториях при помощи `mrpostgres/db` компонентов;
- Переработан принцип получения списков записей из БД с применением сложных фильтров, сортировки и т.д.;
- Изменена логика управления ошибками, переработана инициализация обработки ошибок, которая перенесена в `factory.InitProtoAppErrors()`;
  Переработан `factory.NewErrorHandler`;
- Переработано подключение компонентов `mrsettings` и `mrordering`;

## 2024-09-30
### Added
- Добавлен пакет `tests` используемый в тестировании системы с поддержкой `testcontainers`.
  С помощью `integration.HttpHandlerTester` имеется возможность писать тесты для Http обработчиков,
  каждый из которых будет запускать полноценное приложение с тестовой инфраструктурой.
  При этом инфраструктура может быть развёрнута локально, на тестовом сервере или создаваться
  при помощи библиотеки `testcontainers`;
- Добавлен механизм автоматической миграции БД (с возможностью отключения в конфиге);
- Добавлен отдельный внутренний http сервер (по умолчанию на 8084 порту), где регистрируются
  служебные обработчики: `health`, `system-info`, `metrics` и т.д.;
- К обработчику `health` привязаны пробы БД, redis;
- В обработчик `system-info` добавлен список процессов приложения с их статусами работоспособности;
- Добавлены настройки `traefik` для маппинга портов приложения к доменам. Это позволяет не пробрасывать
  эти порты во вне, тем самым обеспечивается безконфликтный запуск нескольких приложений с одинаковыми портами;
- Добавлены короткие команды в Makefile, обновлена инструкция команд;
- Поправлен `.editorconfig`, добавлены `*.proto`, `*.mk`;
- Поправлены `.env` переменные под новую версию `mrcmd`;

### Changed
- Доработан `config.go` для возможности его использования при написании тестов;
- Переработан запуск приложения, теперь все параллельные сервисы могут запускаться последовательно,
  если работа одного сервиса зависит от работы другого;
- Директория с `./migrations` перенесена в `./app/migrations`;
- Переименовано `Usecase` -> `UseCase`;
- Поле `UpdatedAt` сделано обязательным;
- Отсортированы `imports`;

## 2024-08-03
### Added
- Добавлен дополнительный вывод ошибок для некоторых алгоритмов;

### Changed
- В API документации `Calculations.Algo.PublicAPI.Response.Model.CirculationPackInBox`
  поле `box` переименовано в `fullBox`;
- В API методах, если ни одного значения не было найдено, то возвращается пустой массив, а не nil, как это было ранее;
- Все вспомогательные парсеры перенесены в слой контроллера (в которых происходит перевод значений в СИ);

### Fixed
- Все значения фильтров теперь приводятся к метрическим единицам измерения (СИ);
- При выборке double значений из БД, теперь используется `FilterRangeFloat64`;

## 2024-07-23
### Fixed
- Поправлена и дополнена API документация:
    - в `Calculations.Algo.PublicAPI.Request.Model.RectImposition`
      добавлены пропущенные поля: `disableRotation` и `useMirror`;
    - для v1/calculations/algo/rect/inside-on-outside-max добавлен пропущенный `Request`;
    - operationId сделаны уникальными;

## 2024-07-21
### Added
- Теперь к приложению можно обращаться через локальный домен (с помощью `traefik`).
- Добавлен метод `parallelepiped.Diff()`;

### Changed
- В Open API документации были следующие изменения:
    - добавлены сервера с использованием домена;
    - В объекте `Calculations.Algo.PublicAPI.Response.Model.CirculationPackInBox`:
        - `lastBox` -> `restBox`;
        - `boxVolumeInternal` -> `boxesInnerVolume`;
        - `boxVolumeExternal` -> `boxesVolume`;
        - добавлено поле `boxesWeight`;
    - В объекте `Calculations.Algo.PublicAPI.Response.Model.Box`:
        - добавлены поля `volume`, `innerVolume`;
        - `unusedVolume` -> `unusedVolumePercent`;

### Fixed
- Заменено `PaperLength` -> `PaperWidth`;
- Исправлено `MySQL Group By Sort Style`; 

## 2024-07-14
### Added
- Добавлены образцы файлов с переменными окружения для разных видов разворачивания приложения:
    - `.env.dist` - полный стенд: сервис, запакованный в `docker` и вся необходимая для него инфраструктура;
    - `.env.dist-infra` - только сервис, запакованный в `docker`, инфраструктура разворачивается отдельно;
    - `.env.dist-local` - только сервис, инфраструктура разворачивается отдельно;
- Добавлена переменная `APPX_IMAGES_DIR` для указания внешней директории с файлами изображений;

### Changed
- Добавлены все ENV переменные конфига приложения участвующие в маппинге,
  объединены `.env` и `.env.app`;

### Fixed
- Исправлен баг, при котором по указанному пути не отображалась картинка,
  вместе с этим `ImageProxy.BaseURL` переименовано в `ImageProxy.BasePath`;
- Исправлено добавление/изменения страницы компании (`CompanyPagePostgres.InsertOrUpdate`);
- Добавлено недостающее поле при получении списка коробок;
- Исправлено ошибка получения списка печатных форматов;

## 2024-07-13
### Changed
- Обновлены версии докер образов;
- Переработаны `.env` файлы под различные задачи (локальная сборка, тестовая сборка);

### Fixed
- Исправлен тип `double` в таблицах БД;

## 2024-07-12
### Added
- Добавлены новые API компоненты `App.Field.*`, `App.Field.Measure.*`;
- Добавлено поле `thickness` для коробки;
- Добавлен алгоритм расчёта `PackInBox`;

### Changed
- `laminate_weight` -> `laminate_weight_m2`; 

## 2024-07-07
### Added
- Добавлена документация для модуля сохранения результатов вычислений;

### Changed
- Доработана структура документации, поправлены ссылки;

## 2024-07-06
### Added
- Обновлена документация по командам используемых при разработке;
- Добавлены новые компоненты для `OpenAPI` документации;
- В `App.System.Response.Model.SystemInfo` добавлено поле `environment`;
- Добавлено описание `API` к алгоритмам модуля `Algo` и модулю `QueryHistory`;

### Changed
- Строковые значения `LogLevel` (`logger.level`) приведены к единому стандарту `enum` (теперь они в верхнем регистре);
- Все метрические величины теперь хранятся в БД в системе СИ;
- Для удобства, некоторые метрические величины приходят не в системе СИ,
  но они конвертируются сразу при их получении;
- Файлы с интерфейсами переехали на уровень выше;
- В тестах `defer` заменён на `t.Cleanup`;

## 2024-06-30
### Added
- Подключён планировщик задач `mrworker`.`mrschedule`;
- Подключён компонент `mrsettings` для доступа к произвольным настройкам;
- Подключены `prometheus` метрики для сбора статистики http запросов и работы БД;
- Подключён `mrsentry.Adapter` для отправки ошибок в `sentry`;
- Подключены линтеры с их настройками (`.golangci.yaml`);
- Добавлены комментарии для публичных объектов и методов;
- Добавлена конфигурационная переменная `Environment` для задания рабочего окружения;
- Реструктурированы компоненты типа `RequestParser`, удалены дубликаты;
- Подключено автоматическое определение версии при старте сервиса
  с помощью функции `mrinit.Version()`, но только если явно не указана переменная `APPX_VER`;
- Добавлены комментарии для некоторых структур данных;

### Changed
- Обновлена система формирования ошибок на основе новой версии библиотеки `go-core`:
    - изменён формат создания новых ошибок;
    - объект `AppErrorFactory` заменён на `ProtoAppError` который теперь сам является ошибкой;
- `MimeTypeList` теперь задаётся из `config.yaml`;

### Removed
- Удалена поддержка соединения http сервера по сокету,
  также удалены `ListenTypeSock`, `ListenTypePort`;

## 2024-03-23
### Added
- Добавлены следующие типы ошибок:
    - `FactoryErrPrintFormatNotAvailable`;
    - `FactoryErrLaminateTypeNotAvailable`;
    - `FactoryErrPaperColorNotAvailable`;
    - `FactoryErrPaperFactureNotAvailable`;

### Changed
- Доработан модуль `SubmitForm` и его API документация включая
  библиотеку формирования пользовательских интерфейсов в виде json файлов; 
- В местах использования метода `mrfactory.WithPermission` добавлен `mrfactory.PrepareEachController`;
- `mrserver.NewMiddlewareHttpHandlerAdapter -> mrserver.MiddlewareHandlerAdapter`;
- Доработаны функции типа `factory.registerAdminAPIControllers`, заменены на `createAdminAPIControllers`
  с использованием новой функции `factory.registerControllers`;

### Removed
- Удален метод `IsExist` вместо него теперь используется `FetchStatus`;

## 2024-03-21
### Added
- Добавлены public методы и их API описание для модулей каталога и справочников;

### Changed
- Обновлена структура БД (в том числе поля created_at и updated_at размещены внизу таблицы);
- В `factory.NewRestServer` создание модулей вынесено в методы подобные этому:
  `registerAdminAPIControllers`;
- Переименовано:
    - `ProviderAccountAPI -> ProvidersAPI`;
    - `CompanyPage.PageHead -> PageTitle`;
    - `Custom.Field.Controls.ElementID -> Custom.Field.Controls.FormElementID`;
    - `Custom.Field.Controls.TemplateID -> Custom.Field.Controls.ElementTemplateID`;
    - `Custom.Field.Controls.FormID -> Custom.Field.Controls.SubmitFormID`;
- Добавлено описание ошибки `validator_err_http_url` для `http_url` валидатора;

## 2024-03-19
### Added
- Добавлен `App.Response.Model.SuccessCreatedItemInt32` в API и в `pkg`;
- Добавлены новые типы ошибок (`FactoryErrElementTemplateRequired` и другие);

### Changed
- Перенесены в `pkg` часто используемые сервисом модели:
    - `SuccessCreatedItemResponse`;
    - `ChangeItemStatusRequest`;
    - `MoveItemRequest`;
- Внедрена новая версия библиотеки `go-core`, в связи с этим:
    - в функции `IsAutoCallerOnFunc` изменено условие с использованием `HasCallStack()`;
- В некоторых API методах тип `PUT` преобразован в `PATCH` для более строгого соответствия API спецификации;
- Переработан модуль SubmitForm:
    - идентификатор SubmitForm был заменён с int на uuid;
    - добавлена таблица `submit_forms_compiled` для хранения собранных форм в json формате;
    - добавлен `ActivityStatus`;
    - `FormElement.Required` теперь является необязательным;
    - при создании модуля добавлены дополнительные опции и функции
      `initUnitSubmitFormEnvironment`, `initUnitFormElementEnvironment` чтобы избежать дублирования ресурсов;
- Переименованы методы:
    - `NewFetchParams -> NewSelectParams`;
    - `GetMetaData -> NewOrderMeta`;

### Removed
- Удалён `App.Response.Model.Success`;
- Удалён `App.Response.Model.SuccessModifyItem`;

## 2024-03-16
### Changed
- Все поля БД типа `timestamp` теперь с `with time zone`;
- Заменено `version -> tagVersion`;
- Доработан модуль `ElementTemplate` и его API, добавлена поддержка получения json файла;

## 2024-03-15
### Changed
- Рефакторинг кода:
    - переименование `FactoryErrService* -> FactoryErrUseCase*`, `errService* -> errUseCase*`;
    - переименование интерфейсов `*Service -> *UseCase`;
    - замена методов `LoadOne` на `FetchOne`;
    - методы `Create`, `Insert` теперь возвращают ID записи;
    - схема БД переименована `gscatalog -> sample_catalog`;
- Вся мета информация об изображениях стала необязательной (`imageUrl`, и т.д.);
- Настройки `PageSizeMax` и `PageSizeDefault` вынесены в общие настройки модулей `ModulesSettings.General`;
- Парсер `SortPage` разделён на два: `ListSorter`, `ListPager`;
- Удалены неиспользуемые параметры запросов в каждом из модулей, отсортированы по алфавиту оставшиеся;
- В логгер добавлена поддержка `IsAutoCallerOnFunc`;


- Рефакторинг API документации:
    - Добавлены компоненты:
        - `App.Response.Model.BinaryAnyFile`;
        - `App.Response.Model.BinaryImage`;
        - `App.Response.Model.BinaryMedia`;
        - `App.Response.Model.JsonFile`;
        - `App.Response.Model.SuccessModifyItem`;
        - `App.Response.Model.TextFile.yaml`;
    - Доработка описания фильтрации, сортировки при получении списков записей;
    - Доработка описания ограничений при добавлении/обновлении записей;
    - Для всех модулей поля-идентификаторы описаны как отдельные сущности;

## 2024-02-05
### Changed
- Переименованы:
    - `datetime_created -> created_at`;
    - `datetime_updated -> updated_at`;
    - `modules.Options -> app.Options`;
- Создание модулей переехало в `factory/modules/*`;
- Большинство юнитов было преобразовано в модули, которые объединены доменами;

## 2024-01-30
### Changed
- Внедрён новый интерфейс логгера, добавлен режим трассировки запросов;
- Для многих методов добавлен параметр `ctx context.Context`;
- Заменён устаревший интерфейс `mrcore.EventBox` на `mrsender.EventEmitter`;
- Переименован `ServiceHelper -> UsecaseHelper`;
- Внедрены `mrlib.CallEachFunc`, `CloseFunc` для группового закрытия ресурсов;
- Переименован `CorrelationID` на `X-Correlation-ID`;
- Объекты конфигураций/опций теперь передаются по значению (`*Config -> Config`, `*Options -> Options`);
- Внедрён `oklog/run` для управления одновременным запуском нескольких серверов (http, grpc)
- Добавлены методы для создания и инициализации всех глобальных настроек приложения
  (`CreateAppEnvironment`, `InitAppEnvironment`);
- Теперь модули собираются в рамках отдельных серверов (см. `factory.NewRestServer`);
- Изменены некоторые переменные окружения:
    - удалён `APPX_LOG_PREFIX`;
    - добавлен `APPX_LOG_TIMESTAMP=RFC3339|RFC3339Nano|DateTime|TimeOnly` (формат даты в логах);
    - добавлен `APPX_LOG_JSON=true|false` (вывод логов в json формате);
    - добавлен `APPX_LOG_COLOR=true|false` (использование цветного вывода логов в консоле);
    - переименованы:
        - `APPX_SERVICE_LISTEN_TYPE -> APPX_SERVER_LISTEN_TYPE`;
        - `APPX_SERVICE_LISTEN_SOCK -> APPX_SERVER_LISTEN_SOCK`;
        - `APPX_SERVICE_BIND -> APPX_SERVER_LISTEN_BIND`;
        - `APPX_SERVICE_PORT -> APPX_SERVER_LISTEN_PORT`;

## 2024-01-25
### Added
- Внедрены парсеры на основе интерфейсов `mrserver.RequestParserFile` и
  `mrserver.RequestParserImage` для получения файлов и изображений из `multipart` формы.
    - заменено `mrreq.File -> ht.parser.FormImage`;
    - в `CompanyPageLogoService` изменён тип `mrtype.File -> mrtype.Image`;

### Changed
- Переименовано `ConvertImageMetaToInfo -> ImageMetaToInfoPointer`;

### Removed
- `mrserver.RequestParserPath` удалён вместо него используется
  `mrserver.RequestParserString` и `mrserver.RequestParserParamFunc`;

## 2024-01-22
### Changed
- Расформирован объект `ClientContext` и его одноименный интерфейс, в результате:
    - изменена сигнатура обработчиков с `func(c mrcore.ClientContext)` на `func(w http.ResponseWriter, r *http.Request) error`;
    - с помощью интерфейсов `RequestDecoder`, `ResponseEncoder` можно задавать различные форматы
      принимаемых и отправляемых данных (сейчас реализован только формат `JSON`);
    - запросы обрабатываются встраиваемыми в обработчики объектов `mrparser.*` через интерфейсы:
      `mrserver.RequestParserPath`, `RequestParser`, `RequestParserItemStatus`, `RequestParserKeyInt32`,
      `RequestParserSortPage`, `RequestParserUUID`, `RequestParserValidate`;
    - ответы отправляются встраиваемыми в обработчики объекты `mrresponse.*` через интерфейсы:
      `mrserver.ResponseSender`, `FileResponseSender`, `ErrorResponseSender`;
    - вместо метода `Validate(structRequest any)` используется объект `mrparser.Validator`;
- Произведены следующие замены:
    - `HttpController.AddHandlers -> Handlers() []HttpHandler`
      убрана зависимость контроллера от роутера и секции,
      для установки стандартных разрешений добавлены следующие методы:
      `mrfactory.WithPermission`, `mrfactory.WithMiddlewareCheckAccess`;
    - `ModulesAccess -> AccessControl` (`modules_access -> access_control`) и добавлен интерфейс `mrcore.AccessControl`;
    - `ClientSection -> AppSection` (`client_section -> app_section`) удалена зависимость от `AccessControl`;
- При внедрении новой версии библиотеки `go-core` было заменено:
    - `mrerr.FieldErrorList -> CustomErrorList`;

## 2024-01-19
### Changed
- В БД `enum` типы заменены на `int2` и удалены. Доработано, чтобы `enum` типы сохранялись в виде `int`;
- Код получения файла в обработчике заменён на `mrreq.File`;
- Переименованы методы `checkForm`, `checkLaminate` и подобные в `usecase` на более абстрактный `checkItem`
  (проверяет возможность добавления, сохранения записи);

## 2024-01-17
### Added
- Для каждой секции добавлены настройки `AuthSecret` и `AuthAudience`;
- Добавлены системные обработчики (`RegisterSystemHandlers`);
- Добавлена фильтрация полей справочников (например `Custom.Request.Query.Filter.Density*`);
- Добавлено поле `Config.AppStartedAt` для отслеживания времени запуска сервиса;

### Changed
- Переработана документация `OpenAPI`, все API разделены по секциям, настроена сборка документов.
  Добавлены новые `OpenAPI` компоненты: `App.Field*`, `App.Enum*`, `App.Measure*`,
  `App.Response.Model.*`, и другие;
- Сущности поделены на модули, каждый модуль имеет собственные настройки (`Options`),
  которые назначаются фабрикой сервиса на основе общего конфига;
- Поле `companies_pages.logo_path` заменено на `companies_pages.logo_meta` типа `jsonb`,
  в котором теперь хранится мета информация об изображении;
- В конфиге для всех таймаутов заменён тип `int32` на `time.Duration`;

## 2023-09-21
### Changed
- Обновлены зависимости библиотеки;
- Фиксация зависимостей инфраструктуры;
- Заменён адаптер `*mrpostgres.ConnAdapter` на интерфейс `mrstorage.DbConn`;
- Заменены tabs на пробелы в коде;
- Добавлен модуль страница компании с возможностью загрузки логотипа;
- Добавлен модуль получения файлов;

## 2023-09-17
### Changed
- Исправлены коды возврата в REST API для некоторых методов;
- Все объекты, которые создаются при запуске сервиса перенесены в пакет factory;

### Fixed
- Добавлено закрытие операции получения записей из БД;

## 2023-09-13
### Changed
- Все общие компоненты были вынесены в отдельные проекты: `go-core`, `go-webcore`,
  `go-storage`, `go-components`, в связи с этим были полностью переработаны все связи проекта;
- Обновлены все версии библиотек, от которых зависит проект;

## 2023-08-28
### Added
- Добавлен mrlib.Helper для помощи запуска сервиса;
- Добавлен MiddlewareUserIp() для получения IP текущего пользователя;

### Changed
- Создание соединения postgres перенесено в отдельный класс фабрики;
- Переработаны константы URL адресов методов API;
- Доработано оформление кода логирования событий;
- Изменён интерфейс ClientData, метод SendResponseWithError упразднён;
- Доработана система обработки ошибок в контроллере;
- Метод mrpostgres.ConnAdapter.Close() приведён к стандартному интерфейсу;

## 2023-07-27
### Changed
- Изменена валидация поля article, теперь доступны все символы кроме пробельных;
- Переименованы `ErrHttpRequestEnumLen -> ErrHttpRequestParamLen` и
  `ErrHttpRequestParseEnum -> ErrHttpRequestParseParam` для обобщения;

### Fixed
- Добавлено забытое поле `Sides` в `entity.CatalogPaper`;
- Скорректированы значение валидаторов (min, max, lte)ж
- Добавлен пропущенный фильтр `App.Request.Query.ItemStatuses` в API документацию;
- `*Remove -> *RemoveURL`;
- В методах формирования списков заменён: `client.Query -> client.SqQuery`;

## 2023-07-23
### Add
- Добавлена возможность обновления только указанных полей структуры;
- Добавлено несколько именованных ошибок: `ErrInternalInvalidType`,
  `ErrInternalInvalidData`, `ErrDataContainer`;
- Добавлено `mrentity.EmptynullString` для сохранения NULL значений в БД;

### Changed
- В метод SqUpdate перенесена логика проверки обновления записи;

### Fixed
- Поправлена регулярка для поиска параметров в сообщении, теперь в параметрах могут быть цифры; 

## 2023-07-22
### Add
- Добавлены дополнительные ограничения в таблицы БД;
- Для модельных объектов добавлены константные имена `const ModelName*`,
  которое используется при логировании событий и ошибок связанных с этими компонентами;
- Добавлена структура mrerr.Arg для передачи какой-либо информации в логи;
- Добавлено несколько именованных ошибок: `ErrInternalNilPointer`, `ErrStorageFetchedInvalidData`;
- Добавлены методы работы с БД в связке с `squirrel` (см. `conn_squirrel.go`)

### Changed
- Доработан механизм сортировки элементов, он был оформлен в компонент `ItemOrderer`.
  Его задачи: организовывать порядок следования элементов конкретного списка и
  перемещать эти элементы в рамках этого списка;

### Removed
- Удалён метод NewWithData, вместо него нужно использовать `New(mrerr.Arg{...})`;

## 2023-07-16
### Add
- Добавлен paramName в модуль `form_data`;
- Добавлен компонент для компиляции структуры, на основе которой формируются
  пользовательские интерфейсы для калькуляции продукции;
- В интерфейс `ClientData` добавлен метод `WithContext`;
- Добавлена именованная ошибка `ErrInternalParseData`;

### Changed
- GO_TOOLS утилиты теперь не устанавливаются по умолчанию;
- В ошибке, формируемой во wrapError, теперь выводится информация о том,
  в каком методе разработчика она произошла, а не там, где её породила библиотека ядра;
- Во всех `regexp.MustCompile` добавлены raw strings во избежание двойного экранирования символов;

### Fixed
- Исправлен баг при сохранении id формы в контекст;
- Исправлен баг при добавлении поля формы, проверка `paramName` проходила раньше,
  чем этот параметр заполнялся из шаблона;
- Добавлен omitempty в описание валидации необязательных полей;
- Исправлен синтаксис sql в некоторых запросах;
- Исправлен баг с защитой от бесконечного цикла при формировании ошибки;
- В `FormElementOrderer` добавлены пропущенные проверки ошибок;

## 2023-07-14
### Added
- Доработан проект для его развёртывания утилитой Mrcmd;
- Добавлена ещё одна настройка переменной MRCMD_SHARED_PLUGINS_ENABLED
  для альтернативного варианта разработки;

### Changed
- Переименован пакет `calc-user-data-back-adm -> print-shop-back`;

## 2023-07-10
### Added
- Внедрена библиотека `squirrel` для формирования SQL запросов;

## 2023-07-10
### Added
- Внедрена библиотека `squirrel` для формирования SQL запросов;

## 2023-07-07
### Added
- Произведён рефакторинг механизма обработки ошибок, добавлена фабрика для создания ошибок;

## 2023-05-13
### Changed
- Доработан механизм обработки ошибок, logger и валидация внешних данных;

## 2023-05-12
### Added
- Загружена версия проекта с несколькими реализованными методами API аутентификации;
