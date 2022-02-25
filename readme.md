<h1>Bitmex websocket gateway</h1>

Позволяет подписаться на уведомления об изменении котировок.
Сервис состоит из gateway модуля, provider модуля и брокера сообщений rabbimq.
___

<h3>Gateway</h3>

Предоставляет websocket API для подписки/отписки на изменение котировок.
Получает данные из rabbimq и передает их клиенту.
____

<h3>Provider</h3>

Выполняет роль поставщика данных. Получает значения котировок из API биржи Bitmex
и публикует их в rabbitmq.

____

<h3>TODO List</h3>

- [ ] Тестовое покрытие
- [ ] Авто-реконнекты к rabbitmq и websocket API Bitmex
- [ ] Конфигурация модулей
