# EGTS-сервер

Данный проект включает в себя как библиотеку для кодирования и декодирования EGTS-пакетов, так и непосредственно сам сервер, а также генератор пакетов и API.

## Быстрые ссылки

* [Библиотека](#библиотека);
* [Сервер](#сервер);
* [Установка](#установка);
* [Запуск](#запуск);
* [Запуск в Docker](#запуск-в-docker);
* [Формат конфигурационного файла](#формат-конфигурационного-файла);
* [Развертывание контейнера на тестовом Debian-сервере](#развертывание-контейнера-на-тестовом-debian-сервере);
* [Документация API](./docs/api_documentation.md).

## Библиотека

Библиотека основана на:
* [ГОСТ №54619 от 2011 года](./docs/gost54619-2011.pdf);
* [Приказ №285 Министерства транспорта Российской Федерации от 31.07.2012](./docs/mitrans285.pdf).

Больше информации о протоколе можно найти на [данном ресурсе](https://www.swe-notes.ru/post/protocol-egts/).

**Пример кодирования пакета**:
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

**Пример декодирования пакета**:
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
  postgresql:
    image: postgresql:latest
    container_name: egts_postgresql

  egts:
    image: egts:latest
    container_name: egts_receiver
    ports:
      - "6000:6000"

    volumes:
      - ./configs:/etc/egts-receiver/
```

## Конфигурационный файл

**Формат конфигурационного файла**:
```yaml
host: "127.0.0.1"
provider_id_to_port:
  1: 7000
  2: 7001
api_port: 8000
connection_ttl: 10
log_level: "DEBUG"
log_file_path: "logs/app.log"
log_max_age_days: 366
save_telematics_data_month_start: 5
save_telematics_data_month_end: 9
optimize_geometry_cron_expression: "0 0 2 * * *"
migrations_path: "file://cli/receiver/migrations"

storage:
...
```

**Описание параметров**:
- *host* — адрес;
- *provider_id_to_port* — ассоциативный массив, где ключ — идентификатор провайдера, значение — порт;
- *api_port* — порт API;
- *connection_ttl* — если сервер не получает информацию дольше указанного количества секунд, то соединение закрывается;
- *log_level* — уровень журналирования;
- *log_file_path* — путь до файла с логами;
- *log_max_age_days* — время жизни "старых" файлов с логами;
- *save_telematics_data_month_start* — месяц начала записи телематических данных;
- *save_telematics_data_month_end* — месяц конца записи телематических данных;
- *optimize_geometry_cron_expression* — cron-выражение, определяющее переодичность оптимизации транспортных треков;
- *migrations_path* — путь до директории с файлами миграций;
- *storage* — секция для указания информации о хранилище.

**Описание конфигурационных файлов**:
- *config.yaml*, *config.test.yaml* — для локального запуска;
- *config.docker.yaml* — для локального запуска в Docker;
- *config.docker.test.yaml* — для развертывание на тестовом сервере.

## Развертывание контейнера на тестовом Debian-сервере с помощью GitHub Actions

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