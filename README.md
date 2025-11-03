# backend

install...

enter docker compose up 


### Структура задачи (Task) и её поля: 

* ID (uuid.UUID, обязательное, только для чтения): Уникальный идентификатор задачи, генерируется автоматически. Формат: стандартный UUID , короче набор знаков,циферок и буковок 50400459е94837-an372...

* Title (string, обязательное): Название задачи. Должно быть непустой строкой. (Проверяется тегом binding:"required" и в Validate()).

* Description (string): Описание задачи.

* Status (string, обязательное): Статус задачи. Должен быть одним из предопределенных значений. (Проверяется тегом binding:"required" и в Validate()).  Допустимые значения: "todo", "in_progress", "review", "done".

* KanbanSpace (string, обязательное): Пространство Kanban, к которому относится задача. (Проверяется тегом binding:"required" и в Validate()). Допустимые значения: "backlog", "todo", "in_progress", "review", "done".

* Owner (string, обязательное): Владелец задачи. Должно быть непустой строкой. (Проверяется тегом binding:"required" и в Validate()).

* AssignedTo (указатель на string, опциональное): Пользователь, которому назначена задача. Может быть null.

* Priority (string, опциональное): Приоритет задачи. Если указан, должен быть одним из предопределенных значений. (Проверяется в Validate()).  Допустимые значения: "low", "medium", "high", "urgent". Если не указано, по умолчанию считается "medium" (на уровне логики создания задачи в репозитории).

* DueDate (указатель на time.Time, опциональное): Срок выполнения задачи. Формат времени: RFC3339  Может быть null.

* CreatedAt (time.Time, обязательное, только для чтения): Дата и время создания задачи. Устанавливается автоматически.

* UpdatedAt (time.Time, обязательное, только для чтения): Дата и время последнего обновления задачи. Устанавливается автоматически.

* SubTasks ([]uuid.UUID, только для чтения, в JSON: sub_tasks): Список UUID подзадач, связанных с этой задачей. Заполняется при получении задачи с зависимостями.

* ParentTasks ([]uuid.UUID, только для чтения, в JSON: parent_tasks): Список UUID родительских задач, к которым привязана эта задача. Заполняется при получении задачи с зависимостями.

### Примеры запросов

### Создать задачу
```
curl -X POST http://localhost:8080/api/v1/tasks \
-H "Content-Type: application/json" \
-d '{
"title": "Тестовая задача 4",
"description": "Описание тестовой задачи 4",
"status": "todo",
"kanban_space": "todo",
"owner": "user1",
"priority": "medium"
}'
```

### Получение задачи по id
```
curl -X GET "http://localhost:8080/api/v1//tasks?id= тут айди задачи"
```

### Получить все задачи
```
curl -X GET http://localhost:8080/api/v1//tasks
```

### Обновить задачу
```
curl -X PUT http://localhost:8080/api/v1//tasks \
-H "Content-Type: application/json" \
-d '{
"id": "тут айди задачи",
"title": "Обновленная тестовая задача 4",
"status": "in_progress"
}
```

### Создать вторую задачу (для зависимостей)

```
curl -X POST http://localhost:8080/api/v1//tasks \
-H "Content-Type: application/json" \
-d '{
"title": "Подзадача для задачи 4",
"description": "Это подзадача",
"status": "todo",
"kanban_space": "todo",
"owner": "user1",
"priority": "low"
}'
```

### Добавить зависимость между задачами
```
curl -X POST http://localhost:8080/api/v1//tasks/dependency \
-H "Content-Type: application/json" \
-d '{
"task_id": "тут айди задачи",
"dependent_task_id": "тут айди задачи"
}'
```

### Получить задачу с зависимостями
```
curl -X GET "http://localhost:8080//api/v1/tasks/with-dependencies?id=b39f8904-4b30-4ae2-b3e0-425c3382e928"
```

### Удалить зависимость между задачами
```
curl -X DELETE http://localhost:8080//api/v1/tasks/dependency \
-H "Content-Type: application/json" \
-d '{
"task_id": "тут айди задачи",
"dependent_task_id": "тут айди задачи"
}'
```

