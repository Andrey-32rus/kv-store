# Запуск
docker-compose up -d

# Методы
curl -X PUT -d "data" localhost:8080/kv/foo

curl localhost:8080/kv/foo

curl -X DELETE localhost:8080/kv/foo



curl -X PUT -d "data" https://kv-store-shkm.onrender.com/kv/foo

curl https://kv-store-shkm.onrender.com/kv/foo

curl -X DELETE https://kv-store-shkm.onrender.com/kv/foo