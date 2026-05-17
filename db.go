package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := "user=crm_analyst password=securepassword dbname=crm_automation sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Sprintf("Ошибка подключения к БД: %v", err))
	}

	err = DB.Ping()
	if err != nil {
		panic(fmt.Sprintf("БД недоступна: %v", err))
	}
	fmt.Println("Успешное подключение к PostgreSQL")

	migrationQuery := `
	DROP TABLE IF EXISTS crm_analysis CASCADE;
	DROP TABLE IF EXISTS raw_data CASCADE;

	CREATE TABLE raw_data (
		id SERIAL PRIMARY KEY,
		project_name VARCHAR(255) NOT NULL,
		source_text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE crm_analysis (
		id SERIAL PRIMARY KEY,
		raw_data_id INT REFERENCES raw_data(id) ON DELETE CASCADE,
		recommended_modules TEXT[] NOT NULL,
		extracted_entities TEXT[] NOT NULL,
		complexity_level VARCHAR(50),
		generated_sql TEXT NOT NULL,
		generated_structs TEXT NOT NULL,
		api_documentation TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(migrationQuery)
	if err != nil {
		panic(fmt.Sprintf("Не удалось обновить структуру таблиц: %v", err))
	}
	fmt.Println("Структура БД синхронизирована (добавлены поля кода и API).")
}
