# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

### Prepare

1. Скачать `metricstest` и `random` по ссылке выше для вашей системы. 
2. Переместить в `~/go/bin`, переименовать в `metricstest` и `random`.
3. Для этих файлов сделать
```sh
chmod +x ~/go/bin/metricstest && chmod +x ~/go/bin/random
```
4. Собрать сервер и агент (запуск из корня проекта)
```sh
go build -o ./cmd/server/server ./cmd/server/*.go && go build -o ./cmd/agent/agent ./cmd/agent/*.go
```

### Запуск самих тестов

### Statictest
```sh
go1.22.12 vet -vettool=$(which statictest) ./...
```
### 1
```sh
metricstest -test.v -test.run=^TestIteration1$ \
-binary-path=cmd/server/server
```

### 2
```sh
metricstest -test.v -test.run=^TestIteration2A$ \
-source-path=. \
-agent-binary-path=cmd/agent/agent
```
```sh
metricstest -test.v -test.run=^TestIteration2B$ \
-source-path=. \
-agent-binary-path=cmd/agent/agent
```

### 3
```sh
metricstest -test.v -test.run=^TestIteration3A$ \
-source-path=. \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server
```
```sh
metricstest -test.v -test.run=^TestIteration3B$ \
-source-path=. \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server
```

### 4
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration4$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
```

### 5
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration5$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
```

### 6
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration6$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
```

### 7
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration7$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
```

### 8
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration8$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -server-port=$SERVER_PORT \
  -source-path=.
```

### 9
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration9$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -file-storage-path=$TEMP_FILE \
  -server-port=$SERVER_PORT \
  -source-path=.
```

### 10
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration10A$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
  -server-port=$SERVER_PORT \
  -source-path=.
```
```sh
SERVER_PORT=$(random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(random tempfile)
metricstest -test.v -test.run=^TestIteration10B$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -database-dsn='postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable' \
  -server-port=$SERVER_PORT \
  -source-path=.
```

