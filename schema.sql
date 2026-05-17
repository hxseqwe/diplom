CREATE TABLE IF NOT EXISTS raw_data (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(255) NOT NULL,
    source_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS crm_analysis (
    id SERIAL PRIMARY KEY,
    raw_data_id INT REFERENCES raw_data(id) ON DELETE CASCADE,
    recommended_modules TEXT[] NOT NULL,
    extracted_entities TEXT[] NOT NULL,
    complexity_level VARCHAR(50),
    generated_sql TEXT NOT NULL, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);