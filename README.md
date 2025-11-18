## **Оформление решения**

1. **Скопируйте репозиторий:**

    ```bash
    git clone https://github.com/deplagene/avito-tech-internship.git
    cd avito-backend-trainee
    ```

2. **Пример .env.example**
Посмотрите пример и заполните
.env - сделал публичным.

3. **Запустите Докер-композ:**

    ```bash
    make docker-up
    ```

4. **Примените миграции:**

    ```bash
    make migrate-up
    ```

## **Примеры запросов**

### 1. Создать команду
```bash
curl -X POST http://localhost:8080/team/add \
-H "Content-Type: application/json" \
-d '{
  "team_name": "backend-devs",
  "members": [
    {
      "user_id": "user1",
      "username": "John Doe",
      "is_active": true
    },
    {
      "user_id": "user2",
      "username": "Jane Smith",
      "is_active": true
    },
    {
      "user_id": "user3",
      "username": "Peter Jones",
      "is_active": true
    }
  ]
}'
```

### 2. Получить команду
```bash
curl -X GET "http://localhost:8080/team/get?team_name=backend-devs"
```

### 3. Установить статус активности пользователя
```bash
curl -X POST http://localhost:8080/users/setIsActive \
-H "Content-Type: application/json" \
-d '{
  "user_id": "user1",
  "is_active": false
}'
```

### 4. Создать pull request
```bash
curl -X POST http://localhost:8080/pullRequest/create \
-H "Content-Type: application/json" \
-d '{
  "pull_request_id": "pr123",
  "pull_request_name": "feat: new login flow",
  "author_id": "user1"
}'
```

### 5. Получить pull request для рецензирования
```bash
curl -X GET "http://localhost:8080/users/getReview?user_id=user2"
```

### 6. Переназначить рецензента
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
-H "Content-Type: application/json" \
-d '{
  "pull_request_id": "pr123",
  "old_user_id": "user2"
}'
```

### 7. Выполнить слияние pull request
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
-H "Content-Type: application/json" \
-d '{
  "pull_request_id": "pr123"
}'
```
