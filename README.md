# EGTS-сервер

Данный проект включает в себя как библиотеку для кодирования и декодирования EGTS-пакетов, так и непосредственно сам сервер, а также генератор пакетов и плагины для баз данных.

## Быстрые ссылки

* [Библиотека](#библиотека);
* [Сервер](#сервер);
* [Плагины](#плагины);
* [Установка](#установка);
* [Запуск](#запуск);
* [Запуск в Docker](#запуск-в-docker);
* [Формат конфигурационного файла](#формат-конфигурационного-файла);
* [Развертывание контейнера на тестовом Debian-сервере](#развертывание-контейнера-на-тестовом-debian-сервере).

## Библиотека

Библиотека основана на:
* [ГОСТ №54619 от 2011 года](./docs/gost54619-2011.pdf);
* [Приказ №285 Министерства транспорта Российской Федерации от 31.07.2012](./docs/mitrans285.pdf).

Больше информации о протоколе можно найти на [данном ресурсе](https://www.swe-notes.ru/post/protocol-egts/).

Пример кодирования пакета:
```go
package main 

import (
    "github.com/kuznetsovin/egts-protocol/libs/egts"
    "log"
)

func main() {
    pkg := egts.Package{
    		ProtocolVersion:  1,
    		SecurityKeyID:    0,
    		Prefix:           "00",
    		Route:            "0",
    		EncryptionAlg:    "00",
    		Compression:      "0",
    		Priority:         "11",
    		HeaderLength:     11,
    		HeaderEncoding:   0,
    		FrameDataLength:  3,
    		PacketIdentifier: 137,
    		PacketType:       egts.PtResponsePacket,
    		HeaderCheckSum:   74,
    		ServicesFrameData: &egts.PtResponse{
    			ResponsePacketID: 14357,
    			ProcessingResult: 0,
    		},
    	}
    
    rawPkg, err := pkg.Encode()
	if err != nil {
		log.Fatal(err)
	}
    
    log.Println("Bytes packet: ", rawPkg)
}
```

Пример декодирования пакета:
```go
package main 

import (
    "github.com/kuznetsovin/egts-protocol/libs/egts"
    "log"
)

func main() {
    pkg := []byte{0x01, 0x00, 0x03, 0x0B, 0x00, 0x03, 0x00, 0x89, 0x00, 0x00, 0x4A, 0x15, 0x38, 0x00, 0x33, 0xE8}
    result := egts.Package{}

    state, err := result.Decode(pkg)
    if err != nil {
 		log.Fatal(err)
 	}
    
    log.Println("State: ", state)
    log.Println("Package: ", result)
}
```

## Сервер 

Сервер обрабатывает и по возможности сохраняет всю телематическую информацию из подзаписей типа ```EGTS_SR_POS_DATA```. Если пакет содержит несколько таких подзаписей, то сервер обрабатывает каждую из них.

## Плагины

Для подключения баз данных используются плагины. Каждый плагин должен иметь секцию ```[storage]``` в конфигурационном файле. Также каждый плагин должен реализовывать интерфейс ```Connector```, который представлен ниже.
```go
type Connector interface {
	// setup store connection
	Init(map[string]string) error
	
	// save to store method
	Save(interface{ ToBytes() ([]byte, error) }) error
	
	// close connection with store
	Close() error
}
```

Если конфигурационный файл не имеет секции ```storage```, то будет использоваться стандартный вывод.

Все плагины находятся в [данной директории](/cli/receiver/storage/store).

## Установка

```bash
git clone https://github.com/Daniil11ru/EGTS
cd egts-protocol
make
```

## Запуск

```bash
./bin/receiver -c config.yaml
```

```config.yaml``` – конфигурационный файл.

## Запуск в Docker

Соберите образ:
```bash
make docker
```

Запустите контейнер:

<ul>

<li>

Без указания конфигурационного файла и порта:

```bash
docker run --name egts-receiver egts:latest
```

</li>

<li>

С указанием конфигурационного файла и порта:

```bash
docker run --name egts-receiver -v ./configs:/etc/egts-receiver -p 6000:6000 egts:latest
```

</li>

</ul>

Пример ```docker-compose.yml```:

```yaml
version: '3'

services:
  redis:
    image: redis:latest
    container_name: egts_redis

  egts:
    image: egts:latest
    container_name: egts_receiver
    ports:
      - "6000:6000"

    volumes:
      - ./configs:/etc/egts-receiver/
```

## Формат конфигурационного файла

```yaml
host: "127.0.0.1"
port: "6000"
conn_ttl: 10
log_level: "DEBUG"

storage:
```

Описание параметров:
- *host* – адрес;  
- *port* – порт;
- *conn_ttl* – если сервер не получает информацию дольше указанного количества секунд, то соединение закрывается;
- *log_level* – уровень журналирования;
- *storage* – секция для указания информации о хранилищах.

## Развертывание контейнера на тестовом Debian-сервере

1. Установить *Docker* и *Docker Compose* на тестовом сервере;
2. Завести на тестовом сервере учетную запись *deploy* и наделить правами пользования *Docker* и *Docker Compose*:

	```bash
	adduser --disabled-password --gecos "" deploy
	usermod -aG docker deploy
	```
3. Добавить ключи:
	1. Локально сгенерировать пару ключей:
		```bash
		ssh-keygen -t ed25519 -C "deploy@ci" -f deploy_ci_key
		```
	2. Добавить публичный ключ как ключ развертывания на GitHub (*Settings -> Deploy keys*);
	3. Добавить приватный ключ как ```ACTIONS_SSH_KEY``` в "секреты" репозитория;
	4. Скопировать публичный ключ на тестовый сервер:
		```bash
		su - deploy
		mkdir -p ~/.ssh && chmod 700 ~/.ssh
		echo "ssh-ed25519 AAAAC3NzaC1... mail@domen" >> ~/.ssh/authorized_keys
		chmod 600 ~/.ssh/authorized_keys
		```
	5. Добавить IP тестового сервера как ```TEST_HOST``` в "секреты" репозитория;
	6. Добавить *GitHub Personal Access Token* как ```GHCR_TOKEN``` с правами на чтение, запись и удаление пакетов, а также на чтение кода в "секреты" репозитория.