# Документация API

**Маршруты**
* `GET /api/v1/vehicles`;
* `PATCH /api/v1/vehicles`;
* `GET /api/v1/vehicles/{ID}`;
* `PATCH /api/v1/vehicles/{ID}`;
* `GET /api/v1/vehicles/excel`;
* `GET /api/v1/locations`.

### `GET /api/v1/vehicles`

#### Параметры
| Название          | Описание         |
| ----------------- | ---------------- |
| imei              | IMEI             |
| provider_id       | ID провайдера    |
| moderation_status | Статус модерации |

>Предусмотрено три статуса модерации: `pending` (ожидает модерации), `rejected` (не прошел модерацию), `approved` (одобрен).

#### Пример тела ответа
```json
[
    {
        "oid": 1014463084,
        "name": "О810СМ11",
        "id": 22,
        "imei": "863071014463084",
        "provider_id": 1,
        "moderation_status": "pending"
    },
    {
        "oid": 1014474792,
        "name": "О788НК11",
        "id": 23,
        "imei": "863071014474792",
        "provider_id": 1,
        "moderation_status": "pending"
    },
    {
        "oid": 1014348103,
        "name": "О787НК11",
        "id": 24,
        "imei": "863071014348103",
        "provider_id": 1,
        "moderation_status": "pending"
    },
]
```

<div style="page-break-after: always;"></div>

### `PATCH /api/v1/vehicles`

#### Описание
Обновление по `IMEI`.

#### Пример тела запроса
```json
{
	"name": "О810СМ11",
	"imei": "863071014463084",
	"moderation_status": "approved"
}
```

### `GET /api/v1/vehicles/{ID}`

#### Пример тела ответа
```json
{
	"oid": 1014463084,
	"name": "О810СМ11",
	"id": 22,
	"imei": "863071014463084",
	"provider_id": 1,
	"moderation_status": "pending"
}
```

### `PATCH /api/v1/vehicles/{ID}`

>В рамках данного маршрута `IMEI` не является идентификатором, его передача обновит `IMEI` соответствующего транспорта.

#### Описание
Обновление по `ID`.

#### Пример тела запроса
```json
{
	"name": "О810СМ11",
	"imei": "863071014463084",
	"moderation_status": "approved"
}
```

### `GET /api/v1/vehicles/excel`

#### Параметры
| Название          | Описание         |
| ----------------- | ---------------- |
| provider_id       | ID провайдера    |
| moderation_status | Статус модерации |

>В качестве ответа возвращается бинарное содержимое Excel-файла. Заголовки:
>* `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`;
>* `Content-Disposition: attachment; filename=vehicles.xlsx`.

<div style="page-break-after: always;"></div>

### `GET /api/v1/locations`

#### Параметры
| Название        | Описание                                                               |
| --------------- | ---------------------------------------------------------------------- |
| vehicle_id      | ID транспорта                                                          |
| sent_after      | Время, после которого пакет был отправлен устройством                  |
| sent_before     | Время, до которого пакет был отправлен устройством                     |
| received_after  | Время, после которого пакет был получен сервером                       |
| received_before | Время, до которого пакет был получен сервером                          |
| locations_limit | Максимальное количество местоположений для каждой транспортной единицы |

#### Пример тела ответа
```json
[
    {
        "vehicle_id": 155,
        "locations": [
            {
                "latitude": 63.46989199599947,
                "longitude": 48.84126396124281,
                "altitude": 10,
                "direction": 150,
                "speed": 60,
                "satellite_count": 15,
                "sent_at": "2025-07-01T09:50:43Z",
                "received_at": "2025-07-02T12:42:38Z"
            },
            {
                "latitude": 63.46979332004436,
                "longitude": 48.841283323439136,
                "direction": 150,
                "speed": 60,
                "satellite_count": 15,
                "sent_at": "2025-07-02T12:42:00Z",
                "received_at": "2025-07-02T12:42:56Z"
            }
        ]
    },
    {
        "vehicle_id": 156,
        "locations": [
            {
                "latitude": 62.58923499672423,
                "longitude": 50.86819665759527,
                "altitude": 10,
                "speed": 60,
                "satellite_count": 15,
                "sent_at": "2025-07-01T17:47:01Z",
                "received_at": "2025-07-01T17:50:37Z"
            },
            {
                "latitude": 62.58923499672423,
                "longitude": 50.86819665759527,
                "sent_at": "2025-07-01T21:07:01Z",
                "received_at": "2025-07-01T21:11:37Z"
            }
        ]
    },
]
```