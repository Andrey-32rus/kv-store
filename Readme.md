# Запуск
docker-compose up -d

# Методы
curl -X PUT -d "data" localhost:8080/kv/foo

curl localhost:8080/kv/foo

curl -X DELETE localhost:8080/kv/foo
