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
4. В корне проекта
```sh
chome +x tests.sh
```
5. Для запуска тестов, при запуске будет всегда собираться server и agent

1. Чтобы запустить все тесты
```sh
./test.sh
```
2. Чтобы запустился тест для нужного инкремента
```sh
# ./test.sh {номер инкремента}
./test.sh 4
```
3. Чтобы запустился тест для нескольких инкрементов
```sh
# ./test.sh {с какого инкремента начать} {какой инкремент будет последним}
./test.sh 1-4
```
Будут запущены все тесты от 1 до 4 (1,2,3,4)
