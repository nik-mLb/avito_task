# avito_task

## Запуск 

**команда:** make start
запускает бд postgres на 5433 порту и сам сервер на 8080

## Тесты

**команда:** make test
выполняет тесты и создает папку coverage с cover.out

**команда:** make coverage 
созадет html и показывает покрытие (у меня 78%)

**команда:** make integration-test
выполняет интеграционный тест из условия

## Проблемы

Столкнулся с проблемой, что в какой-то момент на моем интернет соединении при сборке docker compose не подгружались зависимости go(при выполнении go mod download выкидывало ошибку). Но спустя мучения и долгие попытки найти проблему я решил попробовать другой интернет (мобильный) и все получилось!
Собственно из-за этого не успел написать кодген спецификации(

## Дополнительная информация

+ Пример тестирования с помощью apache benchmark можно посмотреть в файле ab_test.txt
+ Пример тестирования с помощью wrk можно посмотреть в файле wrk_test.txt
+ Настроено логгирование с использованием middleware для логгирования каждого запроса, а также логируются все ошибки на каждом уровне(transport, usecase, repository).