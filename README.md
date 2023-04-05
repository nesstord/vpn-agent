# VPN-Agent

## Тестирование бинарника
#### Билд
```sh
./test.sh build <arch>
```
Например:
```sh
./test.sh build amd64
```
По-умолчанию билд для ОС Linux
#### Запуск в контейнере
```sh
./test.sh run
```

## Команды бинарника
#### Установка
```sh
vpn-agent --protocol=shadowsocks install
```
#### Запуск API
```sh
vpn-agent run
```

## TODO
- Добавить поддержку работы в контейнерах с помощью shadowsocks-manager (https://github.com/shadowsocks/shadowsocks-manager)
- Если будут другие протоколы:
  - Добавить поддержку для run-команды разных vpn-протоколов
