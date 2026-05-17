package main

import (
	"fmt"
	"strings"
)

type EntityInfo struct {
	Name        string   `json:"name"`
	TableName   string   `json:"table_name"`
	Fields      []string `json:"fields"`
	ForeignKeys []string `json:"foreign_keys,omitempty"`
}

type AnalysisResult struct {
	Modules    []string `json:"modules"`
	Entities   []string `json:"entities"`
	Complexity string   `json:"complexity"`
	SQLSchema  string   `json:"sql_schema"`
	GoStructs  string   `json:"go_structs"`
	APIDoc     string   `json:"api_doc"`
}

func AnalyzePrimaryData(text string) AnalysisResult {
	lowText := strings.ToLower(text)

	moduleMap := make(map[string]bool)
	entityMap := make(map[string]EntityInfo)

	customerFields := []string{"id SERIAL PRIMARY KEY", "name VARCHAR(255) NOT NULL"}
	if strings.Contains(lowText, "телефон") || strings.Contains(lowText, "номер") {
		customerFields = append(customerFields, "phone VARCHAR(20)")
	}
	if strings.Contains(lowText, "почт") || strings.Contains(lowText, "email") || strings.Contains(lowText, "эмейл") {
		customerFields = append(customerFields, "email VARCHAR(100) UNIQUE")
	}
	if strings.Contains(lowText, "адрес") || strings.Contains(lowText, "доставк") {
		customerFields = append(customerFields, "address TEXT")
	}
	customerFields = append(customerFields, "created_at TIMESTAMP DEFAULT NOW()")

	entityMap["Customers"] = EntityInfo{
		Name:      "Клиент (Customer)",
		TableName: "customers",
		Fields:    customerFields,
	}

	if strings.Contains(lowText, "сделк") || strings.Contains(lowText, "продаж") || strings.Contains(lowText, "лид") {
		moduleMap["Управление продажами (CRM Pipelines)"] = true

		dealFields := []string{"id SERIAL PRIMARY KEY", "customer_id INT", "title VARCHAR(255) NOT NULL"}
		if strings.Contains(lowText, "бюджет") || strings.Contains(lowText, "сумм") || strings.Contains(lowText, "деньг") {
			dealFields = append(dealFields, "amount NUMERIC(12,2) DEFAULT 0.00")
		}
		dealFields = append(dealFields, "status VARCHAR(50) DEFAULT 'new'", "updated_at TIMESTAMP DEFAULT NOW()")

		entityMap["Deals"] = EntityInfo{
			Name:        "Сделка (Deal)",
			TableName:   "deals",
			Fields:      dealFields,
			ForeignKeys: []string{"FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL"},
		}
	}

	if strings.Contains(lowText, "склад") || strings.Contains(lowText, "товар") || strings.Contains(lowText, "остаток") {
		moduleMap["Складской учет (Inventory)"] = true
		entityMap["Products"] = EntityInfo{
			Name:      "Товар (Product)",
			TableName: "products",
			Fields:    []string{"id SERIAL PRIMARY KEY", "sku VARCHAR(100) UNIQUE NOT NULL", "name VARCHAR(255) NOT NULL", "price NUMERIC(10,2)", "stock_quantity INT DEFAULT 0"},
		}
	}

	if moduleMap["Управление продажами (CRM Pipelines)"] && moduleMap["Складской учет (Inventory)"] {
		entityMap["DealProducts"] = EntityInfo{
			Name:      "Содержимое сделки (Deal Items) [M2M]",
			TableName: "deal_products",
			Fields:    []string{"deal_id INT", "product_id INT", "quantity INT DEFAULT 1", "price_at_purchase NUMERIC(10,2)"},
			ForeignKeys: []string{
				"FOREIGN KEY (deal_id) REFERENCES deals(id) ON DELETE CASCADE",
				"FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE",
				"PRIMARY KEY (deal_id, product_id)",
			},
		}
	}

	if strings.Contains(lowText, "звонок") || strings.Contains(lowText, "тикет") || strings.Contains(lowText, "поддержк") {
		moduleMap["Сервисный модуль (Helpdesk)"] = true
		entityMap["Tickets"] = EntityInfo{
			Name:        "Обращение (Ticket)",
			TableName:   "tickets",
			Fields:      []string{"id SERIAL PRIMARY KEY", "customer_id INT", "subject VARCHAR(255) NOT NULL", "description TEXT", "priority VARCHAR(20) DEFAULT 'medium'"},
			ForeignKeys: []string{"FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE"},
		}
	}

	modules := []string{}
	for k := range moduleMap {
		modules = append(modules, k)
	}
	entities := []string{}
	for _, v := range entityMap {
		entities = append(entities, v.Name)
	}

	complexity := "Базовая"
	if len(modules) > 3 {
		complexity = "Сложная (Enterprise)"
	} else if len(modules) > 1 {
		complexity = "Средняя"
	}

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString("-- Сгенерировано CASE-модулем автоматизации\n\n")
	for _, entity := range entityMap {
		sqlBuilder.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", entity.TableName))
		for i, field := range entity.Fields {
			sqlBuilder.WriteString(fmt.Sprintf("    %s", field))
			if i < len(entity.Fields)-1 || len(entity.ForeignKeys) > 0 {
				sqlBuilder.WriteString(",\n")
			} else {
				sqlBuilder.WriteString("\n")
			}
		}
		for i, fk := range entity.ForeignKeys {
			sqlBuilder.WriteString(fmt.Sprintf("    %s", fk))
			if i < len(entity.ForeignKeys)-1 {
				sqlBuilder.WriteString(",\n")
			} else {
				sqlBuilder.WriteString("\n")
			}
		}
		sqlBuilder.WriteString(");\n\n")
	}

	goBuilder := strings.Builder{}
	goBuilder.WriteString("package models\n\nimport \"time\"\n\n")
	for structName, entity := range entityMap {
		goBuilder.WriteString(fmt.Sprintf("// %s представляет таблицу %s\n", structName, entity.TableName))
		goBuilder.WriteString(fmt.Sprintf("type %s struct {\n", structName))
		for _, field := range entity.Fields {
			parts := strings.Fields(field)
			if len(parts) == 0 {
				continue
			}
			fName := strings.Title(strings.ReplaceAll(parts[0], "_id", "ID"))
			fName = strings.ReplaceAll(fName, "Id", "ID")
			fType := "string"

			if strings.Contains(field, "SERIAL") || strings.Contains(field, "INT") {
				fType = "int"
			}
			if strings.Contains(field, "NUMERIC") {
				fType = "float64"
			}
			if strings.Contains(field, "TIMESTAMP") || strings.Contains(field, "DATE") {
				fType = "time.Time"
			}

			goBuilder.WriteString(fmt.Sprintf("\t%-12s %-10s `json:\"%s\"`\n", fName, fType, parts[0]))
		}
		goBuilder.WriteString("}\n\n")
	}

	apiBuilder := strings.Builder{}
	apiBuilder.WriteString("### Спецификация REST API\n\n")
	for _, entity := range entityMap {
		apiBuilder.WriteString(fmt.Sprintf("#### Сущность: %s\n", entity.TableName))
		apiBuilder.WriteString(fmt.Sprintf("- `GET /api/v1/%s` — Получить список\n", entity.TableName))
		apiBuilder.WriteString(fmt.Sprintf("- `POST /api/v1/%s` — Создать новую запись\n", entity.TableName))
		apiBuilder.WriteString(fmt.Sprintf("- `DELETE /api/v1/%s/:id` — Удалить запись\n\n", entity.TableName))
	}

	return AnalysisResult{
		Modules:    modules,
		Entities:   entities,
		Complexity: complexity,
		SQLSchema:  sqlBuilder.String(),
		GoStructs:  goBuilder.String(),
		APIDoc:     apiBuilder.String(),
	}
}
