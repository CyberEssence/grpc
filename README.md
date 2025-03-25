# Тестовое задание на golang разработчика

Инструкция по запуску приложения:
1) перейти в директорию test
2) создать 2 бд в pgadmin (утилита postgresql) с названиями auth_service и call_service (предварительно должен быть установлен PostgreSQL 17)
2) собрать docker образ командой docker-compose.exe up --build (предварительно должен быть установлен docker desktop на винде)
3) далее выполнять запросы через grpcui/curl

Запросы к регистрации/логину можно отправлять с помощью утилиты grpcui после запуска образа docker с приложением

Ссылка на программу: https://github.com/fullstorydev/grpcui

Запуск программы: 

grpcui.exe -plaintext localhost:50051

Запросы CRUD к заявкам можно выполнить через curl (cmd)

Пример некоторых запросов:

curl -X POST http://localhost:8080/calls -H "Content-Type: application/json" -H "Authorization: Bearer <YOUR_BEARER_TOKEN>" -d "{\"client_name\": \"John Doe\", \"phone_number\": \"+1234567890\", \"description\": \"Issue with service\"}"

curl -X GET http://localhost:8080/calls -H "Authorization: Bearer YOUR_TOKEN"

curl -X PATCH http://localhost:8080/calls/<CALL_ID>/status -H "Content-Type: application/json" -H "Authorization: Bearer <YOUR_BEARER_TOKEN>" -d "{\"status\": \"закрыта\"}"

Также можно запустить тесты, перейдя по пути test\call-service\internal\handler командой go test
