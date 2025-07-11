# Запуск
docker build -t kvstore-app .
docker run -p 8080:8080 kvstore-app

# Методы
curl -X PUT -d "data" https://kv-store-shkm.onrender.com/kv/foo

curl https://kv-store-shkm.onrender.com/kv/foo

curl -X DELETE https://kv-store-shkm.onrender.com/kv/foo